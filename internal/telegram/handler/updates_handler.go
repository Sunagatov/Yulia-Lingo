package handler

import (
	"Yulia-Lingo/internal/telegram/setup"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleBotUpdates() error {
	bot, err := setup.GetTelegramBot()
	if err != nil {
		return fmt.Errorf("app wosn't connect to telegram bot, err: %v", err)
	}

	botUpdates := bot.ListenForWebhook("/")
	for telegramBotUpdate := range botUpdates {
		handleBotUpdate(telegramBotUpdate)
	}
	return nil
}

func handleBotUpdate(botUpdate tgbotapi.Update) {
	if botUpdate.Message != nil {
		HandleMessageFromUser(botUpdate)
	} else if botUpdate.CallbackQuery != nil {
		HandleCallbackQuery(botUpdate)
	}
}
