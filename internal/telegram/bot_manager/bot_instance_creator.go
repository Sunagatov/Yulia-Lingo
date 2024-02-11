package bot_manager

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
)

func CreateTelegramBot() (*tgbotapi.BotAPI, error) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		return nil, fmt.Errorf("failed to read TELEGRAM_BOT_TOKEN from the environment variabless")
	}
	return tgbotapi.NewBotAPI(botToken)
}
