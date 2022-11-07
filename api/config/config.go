package config

import (
	"os"
)

type BotConfig struct {
	TelegramToken, WebhookSite, Port, GithubGQLToken string
}

func LoadBotConfig() *BotConfig {
	if os.Getenv("FROM_FILE") == "1" {
		// TODO: implement
		return nil
	} else {
		return &BotConfig{
			TelegramToken:  os.Getenv("TELEGRAM_BOT_TOKEN"),
			WebhookSite:    os.Getenv("TELEGRAM_BOT_SITE_URL"),
			Port:           os.Getenv("TELEGRAM_BOT_PORT"),
			GithubGQLToken: os.Getenv("GITHUB_GQL_TOKEN"),
		}
	}
}
