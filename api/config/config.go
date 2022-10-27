package config

import (
	"os"
)

type BotConfig struct {
	Token, WebhookSite, Port string
}

func LoadBotConfig() (*BotConfig) {
	if os.Getenv("FROM_FILE") == "1" {
		// TODO: implement 
		return nil
	} else {
		return &BotConfig{
			Token: os.Getenv("TELEGRAM_BOT_TOKEN"),
			WebhookSite: os.Getenv("TELEGRAM_BOT_SITE_URL"),
			Port: os.Getenv("TELEGRAM_BOT_PORT"),
		}
	}
}
