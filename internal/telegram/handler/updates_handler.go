package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleBotUpdates(bot *tgbotapi.BotAPI) {
	updateEndpoint := "/" + bot.Token
	botUpdates := bot.ListenForWebhook(updateEndpoint)

	for telegramBotUpdate := range botUpdates {
		handleBotUpdate(bot, telegramBotUpdate)
	}
}

func handleBotUpdate(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update) {
	if botUpdate.Message != nil {
		HandleMessageFromUser(bot, botUpdate)
	} else if botUpdate.CallbackQuery != nil {
		HandleCallbackQuery(bot, botUpdate)
	}
}
