package main

import (
	"fmt"
	"log"
	"os"

	"github.com/chofnar/release-bot/api/consts"
	"github.com/chofnar/release-bot/api/messages"
	"github.com/mymmrac/telego"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	site := os.Getenv("TELEGRAM_BOT_SITE_URL")
	port := os.Getenv("TELEGRAM_BOT_PORT")

	bot, err := telego.NewBot(token, telego.WithDefaultDebugLogger())
	if err != nil {
		log.Println(err)

		panic(err)
	}

	_ = bot.SetWebhook(&telego.SetWebhookParams{
		URL: "https://" + site + ":443" + "/bot/" + token,
	})

	info, _ := bot.GetWebhookInfo()
	fmt.Printf("Webhook info: %+v\n", info)

	updates, _ := bot.UpdatesViaWebhook("/bot/" + token)

	err = bot.StartListeningForWebhook("0.0.0.0" + ":" + port)

	defer func() {
		_ = bot.StopWebhook()
	}()

	for update := range updates {
		fmt.Printf("Update: %+v\n", update)
		chatID := update.Message.Chat.ID
		// command
		if update.Message.Text != "" {
			switch update.Message.Text {
			case "/start":
				bot.SendMessage(messages.StartMessage(chatID))

			case "/about":
				bot.SendMessage(messages.AboutMessage(chatID))

			default: 
				bot.SendMessage(messages.UnknownCommandMessage(chatID))
				bot.SendMessage(messages.StartMessage(chatID))
			}
		} else {
			switch update.CallbackQuery.Data {
			case consts.SeeAllCallback:
				//bot.EditMessageText()
				//bot.EditMessageReplyMarkup()
				bot.SendMessage(messages.SeeAllReposMessage(chatID))

			case consts.AddCallback:
				//bot.EditMessageText()
				//bot.EditMessageReplyMarkup()
				bot.SendMessage(messages.AddRepoMessage(chatID))

			default:
				break
			}
		}
	}
}
