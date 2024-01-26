package handler

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageHandler struct {
	TelegramBot *tgbotapi.BotAPI
	Update      *tgbotapi.Update
}

func LaunchTelegramBot(telegramBot *tgbotapi.BotAPI) {
	updateEndpoint := "/" + telegramBot.Token
	telegramBotUpdates := telegramBot.ListenForWebhook(updateEndpoint)

	for telegramBotUpdate := range telegramBotUpdates {
		processUpdate(telegramBot, &telegramBotUpdate)
	}
}

func processUpdate(telegramBot *tgbotapi.BotAPI, update *tgbotapi.Update) {

	messageHandler := MessageHandler{
		TelegramBot: telegramBot,
		Update:      update,
	}

	if update.CallbackQuery != nil {
		messageHandler.CallbackQuery()

	} else if update.Message.IsCommand() {
		log.Printf("User %s touched the command", update.Message.From.UserName)

	} else if update.Message.Text != "" {
		log.Printf("User %s touched the text message", update.Message.From.UserName)
		messageHandler.TextMessage()
	}
}
