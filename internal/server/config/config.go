package config

import (
	"os"
	"strconv"
)

type BotConfig struct {
	TelegramToken, WebhookSite, WebhookPort, Port, GithubGQLToken, ResetWebhookUrl string
	Limit                                                                          int
}

func LoadBotConfig() *BotConfig {
	if os.Getenv("FROM_FILE") == "1" {
		// TODO: implement
		return nil
	} else {
		limit, err := strconv.Atoi(os.Getenv("LIMIT"))
		if err != nil || limit == 0 {
			panic("Limit cannot be read or is 0!")
		}

		return &BotConfig{
			TelegramToken:   os.Getenv("TELEGRAM_BOT_TOKEN"),
			WebhookSite:     os.Getenv("TELEGRAM_BOT_SITE_URL"),
			Port:            os.Getenv("PORT"),
			WebhookPort:     os.Getenv("WEBHOOK_PORT"),
			GithubGQLToken:  os.Getenv("GRAPHQL_TOKEN"),
			ResetWebhookUrl: os.Getenv("RESET_WEBHOOK_URL"),
			Limit:           limit,
		}
	}
}
