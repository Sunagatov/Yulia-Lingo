package telegram

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func CreateNewTelegramBot() (*tgbotapi.BotAPI, error) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		return nil, fmt.Errorf("no TELEGRAM_BOT_TOKEN provided in environment variables")
	}

	telegramBot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to authorize Telegram telegramBot with token %s: %v", botToken, err)
	}
	telegramBot.Debug = true

	log.Printf("Authorized on Telegram account: %s", telegramBot.Self.UserName)
	return telegramBot, nil
}
