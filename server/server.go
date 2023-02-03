package server

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"

	"github.com/chofnar/release-bot/database"
	databaseLoader "github.com/chofnar/release-bot/database/loader"
	"github.com/chofnar/release-bot/server/behaviors"
	botConfig "github.com/chofnar/release-bot/server/config"
	"github.com/chofnar/release-bot/server/consts"
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
	linkRegex, _ := regexp.Compile("(?:https://)github.com[:/](.*)[:/](.*)")
	directRegex, _ := regexp.Compile("(.*)[/](.*)")
	behaviorHandler := behaviors.BehaviorHandler{
		Bot:         bot,
		LinkRegex:   linkRegex,
		DirectRegex: directRegex,
		GQLClient:   githubGQLClient,
		DB:          db,
	}
	awaitingAddRepo := map[int64]struct{}{}
	type void struct{}
	var set void

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
					err = behaviorHandler.Start(chatID)
					if err != nil {
						logger.Error(err)
					}

				case "/about":
					err = behaviorHandler.About(chatID)
					if err != nil {
						logger.Error(err)
					}

				default:
					if _, ok := awaitingAddRepo[chatID]; !ok {
						err = behaviorHandler.UnknownCommand(chatID)
						if err != nil {
							logger.Error(err)
						}
					} else {
						err = behaviorHandler.SentRepo(update.Message.Text, update.Message.MessageID, chatID)
						if err != nil {
							logger.Error(err)
						}
					}
				}
			} else {
				chatID := update.CallbackQuery.Message.Chat.ID
				messageID := update.CallbackQuery.Message.MessageID

				switch update.CallbackQuery.Data {
				case consts.SeeAllCallback:
					err := behaviorHandler.SeeAll(chatID, messageID)
					if err != nil {
						logger.Error(err)
					}

				case consts.AddCallback:
					err = behaviorHandler.Add(chatID, messageID)
					if err != nil {
						logger.Error(err)
					}
					awaitingAddRepo[chatID] = set

				case consts.MenuCallback:
					err := behaviorHandler.Menu(chatID, messageID)
					if err != nil {
						logger.Error(err)
					}
					delete(awaitingAddRepo, chatID)

				//TODO: can probably just use the name now
				// name hash callback, delete
				default:
					err = behaviorHandler.DeleteRepo(chatID, messageID, update.CallbackQuery.Data)
					if err != nil {
						logger.Error(err)
					}
				}
			}
		}
	}
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
