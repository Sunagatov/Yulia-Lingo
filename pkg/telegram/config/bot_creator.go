package telegram

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func CreateNewTelegramBot() *tgbotapi.BotAPI {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("No TELEGRAM_BOT_TOKEN provided in environment variables")
	}

	telegramBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Failed to authorize Telegram telegramBot with token %s: %v", botToken, err)
	}
	telegramBot.Debug = true

	log.Printf("Authorized on Telegram account: %s", telegramBot.Self.UserName)
	return telegramBot
}
