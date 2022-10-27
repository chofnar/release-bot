package main

import (
	"fmt"
	"log"

	botConfig "github.com/chofnar/release-bot/api/config"
	"github.com/chofnar/release-bot/database"
	databaseLoader "github.com/chofnar/release-bot/database/loader"
	"github.com/chofnar/release-bot/api/consts"
	"github.com/chofnar/release-bot/api/messages"
	"github.com/mymmrac/telego"

	"go.uber.org/zap"
)

func Initialize(logger zap.SugaredLogger) (*botConfig.BotConfig, *database.Database) {
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
	botConf, _ := Initialize(*logger)

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
		fmt.Printf("Update: %+v\n", update)
		chatID := update.Message.Chat.ID
		messageID := update.Message.MessageID
		var err error
		
		// command
		if update.Message.Text != "" {
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
				_, err = bot.SendMessage(messages.UnknownCommandMessage(chatID))
				if err != nil {
					logger.Error(err)
				}

				_, err = bot.SendMessage(messages.StartMessage(chatID))
				if err != nil {
					logger.Error(err)
				}
			}
		} else {
			switch update.CallbackQuery.Data {
			case consts.SeeAllCallback:
				_, err = bot.EditMessageText(messages.SeeAllReposMessage(chatID, messageID))
				if err != nil {
					logger.Error(err)
				}
				
				_, err = bot.EditMessageReplyMarkup(messages.AddRepoMarkup(chatID, messageID))
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

			default:
				break
			}
		}
	}
}
