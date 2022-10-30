package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	botConfig "github.com/chofnar/release-bot/api/config"
	"github.com/chofnar/release-bot/api/consts"
	"github.com/chofnar/release-bot/api/messages"
	"github.com/chofnar/release-bot/api/repo"
	"github.com/chofnar/release-bot/database"
	databaseLoader "github.com/chofnar/release-bot/database/loader"
	"github.com/mymmrac/telego"

	"go.uber.org/zap"
)

func Initialize(logger zap.SugaredLogger) (*botConfig.BotConfig, database.Database) {
	conf := botConfig.LoadBotConfig()
	db := databaseLoader.GetDatabase(logger)
	return conf, db
}

func main() {
	unsugared, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	logger := unsugared.Sugar()
	botConf, db := Initialize(*logger)

	awaitingAddRepo := map[int64]struct{}{}
	type void struct{}
	var set void

	linkRegex, _ := regexp.Compile("(?:https://)github.com[:/](.*)[:/](.*)")
	directRegex, _ := regexp.Compile("(.*)[/](.*)")

	bot, err := telego.NewBot(botConf.Token, telego.WithDefaultDebugLogger())
	if err != nil {
		logger.Error(err)

		panic(err)
	}

	_ = bot.SetWebhook(&telego.SetWebhookParams{
		URL: "https://" + botConf.WebhookSite + ":443" + "/bot/" + botConf.Token,
	})

	info, _ := bot.GetWebhookInfo()
	fmt.Printf("Webhook info: %+v\n", info)

	updates, _ := bot.UpdatesViaWebhook("/bot/" + botConf.Token)

	err = bot.StartListeningForWebhook("0.0.0.0" + ":" + botConf.Port)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = bot.StopWebhook()
	}()

	for update := range updates {
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
					// TODO: input validation, add to db
					owner, repoName, valid := validateInput(update.Message.Text, linkRegex, directRegex)
					if valid {
						_, err = validateAndBuildRepo(owner, repoName)
						if err != nil {
							logger.Error(err)
							// TODO: invalid repo
						}
						err = db.AddRepo(nil, nil)
						if err != nil {
							logger.Error(err)
							// TODO: warn that something went wrong
						}
					} else {
						// TODO: implement
						break
					}
				}
			}
		} else {
			chatID := update.CallbackQuery.Message.Chat.ID
			messageID := update.CallbackQuery.Message.MessageID

			switch update.CallbackQuery.Data {
			case consts.SeeAllCallback:
				_, err = bot.EditMessageText(messages.SeeAllReposMessage(chatID, messageID))
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

			case consts.AddCallback:
				_, err = bot.EditMessageText(messages.AddRepoMessage(chatID, messageID))
				if err != nil {
					logger.Error(err)
				}

				_, err = bot.EditMessageReplyMarkup(messages.AddRepoMarkup(chatID, messageID))
				if err != nil {
					logger.Error(err)
				}

				awaitingAddRepo[chatID] = set

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

// TODO: replace with GraphQL
func validateAndBuildRepo(owner, name string) (repo.Repo, error) {
	resp, err := http.Get("https://api.github.com/repos/" + owner + "/" + name + "/releases")
	if err != nil {
		return repo.Repo{}, err
	}

	resultingRepo := repo.HelperRepo{}
	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Print(string(bodyBytes))

	err = json.NewDecoder(resp.Body).Decode(&resultingRepo)
	if err != nil {
		return repo.Repo{}, err
	}

	return repo.Repo{}, nil
}
