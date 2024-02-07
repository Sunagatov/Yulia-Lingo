package handler

import (
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleBotUpdates(bot *tgbotapi.BotAPI, db *sql.DB) {
	updateEndpoint := "/" + bot.Token
	botUpdates := bot.ListenForWebhook(updateEndpoint)

	for telegramBotUpdate := range botUpdates {
		handleBotUpdate(bot, db, telegramBotUpdate)
	}
}

func handleBotUpdate(bot *tgbotapi.BotAPI, db *sql.DB, botUpdate tgbotapi.Update) {
	if botUpdate.Message != nil {
		HandleMessageFromUser(bot, db, botUpdate)
	} else if botUpdate.CallbackQuery != nil {
		HandleCallbackQuery(bot, db, botUpdate)
	}
}
