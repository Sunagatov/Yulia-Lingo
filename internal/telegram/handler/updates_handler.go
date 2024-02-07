package handler

import (
	"Yulia-Lingo/internal/telegram"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleBotUpdates() error {
	bot, err := telegram.GetTelegramBot()
	if err != nil {
		return fmt.Errorf("app wosn't connect to telegram bot, err: %v", err)
	}
	botUpdates := bot.ListenForWebhook("/")
	for telegramBotUpdate := range botUpdates {
		handleBotUpdate(bot, telegramBotUpdate)
	}
	return nil
}

func handleBotUpdate(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update) {
	if botUpdate.Message != nil {
		HandleMessageFromUser(bot, botUpdate)
	} else if botUpdate.CallbackQuery != nil {
		HandleCallbackQuery(bot, botUpdate)
	}
}
