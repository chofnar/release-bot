package api

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"

	botConfig "github.com/chofnar/release-bot/api/config"
	"github.com/chofnar/release-bot/api/consts"
	"github.com/chofnar/release-bot/api/messages"
	"github.com/chofnar/release-bot/api/repo"
	"github.com/chofnar/release-bot/database"
	databaseLoader "github.com/chofnar/release-bot/database/loader"
	"github.com/chofnar/release-bot/errors"
	"github.com/hasura/go-graphql-client"
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

func Initialize(logger zap.SugaredLogger) (*botConfig.BotConfig, database.Database) {
	conf := botConfig.LoadBotConfig()
	db := databaseLoader.GetDatabase(logger)
	return conf, db
}

func Start() {
	unsugared, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	logger := unsugared.Sugar()
	botConf, db := Initialize(*logger)

	bot, err := telego.NewBot(botConf.TelegramToken, telego.WithLogger(logger))
	if err != nil {
		logger.Error(err)

		panic(err)
	}

	// can't get ngrok to forward properly without this
	if os.Getenv("STAGING") == "TRUE" {
		err = bot.SetWebhook(&telego.SetWebhookParams{
			URL: "https://" + botConf.WebhookSite + "/bot/" + botConf.TelegramToken,
		})
		if err != nil {
			logger.Error(err)

			panic(err)
		}
	}

	updates, err := bot.UpdatesViaWebhook("/bot/" + botConf.TelegramToken)
	if err != nil {
		logger.Error(err)

		panic(err)
	}

	err = bot.StartListeningForWebhook("0.0.0.0" + ":" + botConf.Port)
	if err != nil {
		logger.Error(err)

		panic(err)
	}

	defer func() {
		_ = bot.StopWebhook()
	}()

	// Create Github GraphQL token
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: botConf.GithubGQLToken},
	)

	http.HandleFunc("/", home)
	// updater listener, called from cron
	http.HandleFunc("/updateRepos", updateRepos)
	go http.ListenAndServe(":"+botConf.Port, nil)

	httpClient := oauth2.NewClient(context.Background(), src)
	githubGQLClient := graphql.NewClient("https://api.github.com/graphql", httpClient)

	ctx := context.Background()
	nctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	go updateLoop(ctx, updates, bot, db, githubGQLClient, *logger)
	defer stop()
	<-nctx.Done()
}

func updateLoop(ctx context.Context, updates <-chan telego.Update, bot *telego.Bot, db database.Database, githubGQLClient *graphql.Client, logger zap.SugaredLogger) {
	type void struct{}
	var set void
	linkRegex, _ := regexp.Compile("(?:https://)github.com[:/](.*)[:/](.*)")
	directRegex, _ := regexp.Compile("(.*)[/](.*)")

	awaitingAddRepo := map[int64]struct{}{}

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			var err error

			// command
			if update.Message != nil {
				chatID := update.Message.Chat.ID

				switch update.Message.Text {
				case "/start":
					_, err = bot.SendMessage(messages.StartMessage(chatID))
					if err != nil {
						logger.Error(err)
					}

				case "/about":
					_, err = bot.SendMessage(messages.AboutMessage(chatID))
					if err != nil {
						logger.Error(err)
					}

				default:
					if _, ok := awaitingAddRepo[chatID]; !ok {
						_, err = bot.SendMessage(messages.UnknownCommandMessage(chatID))
						if err != nil {
							logger.Error(err)
						}

						_, err = bot.SendMessage(messages.StartMessage(chatID))
						if err != nil {
							logger.Error(err)
						}
					} else {
						owner, repoName, valid := validateInput(update.Message.Text, linkRegex, directRegex)
						if valid {
							repoToAdd, err := validateAndBuildRepo(owner, repoName, githubGQLClient)
							hasReleases := true
							if err != nil {
								if err == errors.ErrNoReleases {
									hasReleases = false

								} else {
									logger.Error(err)

									_, err = bot.SendMessage(messages.RepoNotFoundMessage(chatID))
									if err != nil {
										logger.Error(err)
									}
									break
								}
							}
							exists, err := db.CheckExisting(fmt.Sprint(chatID), repoToAdd.RepoID)
							if err != nil {
								logger.Error(err)

							}

							if exists {
								_, err = bot.SendMessage(messages.AlreadyExistsMessage(chatID, update.Message.MessageID))
								if err != nil {
									logger.Error(err)
								}
								break
							}

							err = db.AddRepo(fmt.Sprint(chatID), &repoToAdd)
							if err != nil {
								logger.Error(err)
							}

							if hasReleases {
								_, err = bot.SendMessage(messages.SuccesfullyAddedRepoMessage(chatID))
								if err != nil {
									logger.Error(err)
								}
							} else {
								_, err = bot.SendMessage(messages.SuccesfullyAddedRepoWithoutReleasesMessage(chatID))
								if err != nil {
									logger.Error(err)
								}
							}

						} else {
							_, err = bot.SendMessage(messages.InvalidRepoMessage(chatID))
							if err != nil {
								logger.Error(err)
							}
						}
					}
				}
			} else {
				chatID := update.CallbackQuery.Message.Chat.ID
				messageID := update.CallbackQuery.Message.MessageID

				switch update.CallbackQuery.Data {
				case consts.SeeAllCallback:
					markup, err := messages.SeeAllReposMarkup(chatID, messageID, &db)
					if err != nil {
						logger.Error(err)
					}
					if err == errors.ErrNoRepos {
						_, err = bot.EditMessageText(messages.SeeAllReposButNoneFoundMessage(chatID, messageID, *markup.ReplyMarkup))
						if err != nil {
							logger.Error(err)
						}
					} else {
						_, err = bot.EditMessageText(messages.SeeAllReposMessage(chatID, messageID, *markup.ReplyMarkup))
						if err != nil {
							logger.Error(err)
						}
					}

				case consts.AddCallback:
					_, err = bot.EditMessageText(messages.AddRepoMessage(chatID, messageID))
					if err != nil {
						logger.Error(err)
					}

					awaitingAddRepo[chatID] = set

				case consts.MenuCallback:
					_, err = bot.EditMessageText(messages.EditedStartMessage(chatID, messageID))
					if err != nil {
						logger.Error(err)
					}
					delete(awaitingAddRepo, chatID)

				//TODO: can probably just use the name now
				// name hash callback, delete
				default:
					err = messages.DeleteRepo(chatID, update.CallbackQuery.Data, &db)
					if err != nil {
						logger.Error(err)
					}

					markup, err := messages.SeeAllReposMarkup(chatID, messageID, &db)
					if err != nil {
						logger.Error(err)
					}

					_, err = bot.EditMessageReplyMarkup(markup)
					if err != nil {
						logger.Error(err)
					}
				}
			}
		}
	}
}

func validateInput(message string, linkRegex, directRegex *regexp.Regexp) (owner, repo string, isValid bool) {
	if result := linkRegex.FindStringSubmatch(message); result != nil {
		owner, repo = result[1], result[2]
		isValid = true
		return
	}

	if result := directRegex.FindString(message); result != "" {
		splitResult := strings.Split(message, "/")
		owner, repo = splitResult[0], splitResult[1]
		isValid = true
		return
	}

	return
}

func validateAndBuildRepo(owner, name string, client *graphql.Client) (repo.Repo, error) {
	variables := map[string]interface{}{
		"name":  name,
		"owner": owner,
	}

	var getRepoQuery struct {
		Repository struct {
			ID    string
			URL   string
			Name  string
			Owner struct {
				Login string
			}
			Releases struct {
				Nodes []struct {
					TagName string
					ID      string
				}
			} `graphql:"releases(first: 1)"`
		} `graphql:"repository(name: $name, owner: $owner)"`
	}

	err := client.Query(context.TODO(), &getRepoQuery, variables)
	if err != nil {
		return repo.Repo{}, err
	}

	if len(getRepoQuery.Repository.Releases.Nodes) != 0 {
		return repo.Repo{
			RepoID: getRepoQuery.Repository.ID,
			Name:   getRepoQuery.Repository.Name,
			Owner:  getRepoQuery.Repository.Owner.Login,
			Link:   getRepoQuery.Repository.URL,
			Release: repo.Release{
				CurrentReleaseTagName: getRepoQuery.Repository.Releases.Nodes[0].TagName,
				CurrentReleaseID:      getRepoQuery.Repository.Releases.Nodes[0].ID,
			},
		}, nil
	}

	return repo.Repo{
		RepoID: getRepoQuery.Repository.ID,
		Name:   getRepoQuery.Repository.Name,
		Owner:  getRepoQuery.Repository.Owner.Login,
		Link:   getRepoQuery.Repository.URL,
		Release: repo.Release{
			CurrentReleaseTagName: "",
			CurrentReleaseID:      "",
		},
	}, errors.ErrNoReleases
}

func updateRepos(w http.ResponseWriter, r *http.Request) {
	// TODO: dump db

	// TODO: check for new releases

	// TODO: send messages

	io.WriteString(w, "Updated")
}

func home(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Something's working alright")
}
