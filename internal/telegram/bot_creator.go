package telegram

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	bot *tgbotapi.BotAPI
	err error
)

func CreateNewTelegramBot() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatalf("no TELEGRAM_BOT_TOKEN provided in environment variables")
	}

	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("failed to authorize Telegram telegramBot with token %s: %v", botToken, err)
	}
	bot.Debug = true

	log.Printf("Authorized on Telegram account: %s", bot.Self.UserName)
}

func GetTelegramBot() (*tgbotapi.BotAPI, error) {
	return bot, err
}
