package bot_manager

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
)

func CreateTelegramBot() (*tgbotapi.BotAPI, error) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatalf("No TELEGRAM_BOT_TOKEN provided in environment variables")
	}
	return tgbotapi.NewBotAPI(botToken)
}
