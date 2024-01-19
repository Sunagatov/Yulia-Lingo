package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func listenForUpdates(bot *tgbotapi.BotAPI) {
	updateEndpoint := "/" + bot.Token
	updates := bot.ListenForWebhook(updateEndpoint)
	for update := range updates {
		if update.Message == nil {
			continue
		}

		userMessage := update.Message.Text

		responseMessage := "Thank you for your message: " + userMessage

		newMessage := tgbotapi.NewMessage(update.Message.Chat.ID, responseMessage)

		_, sendingMessageError := bot.Send(newMessage)
		if sendingMessageError != nil {
			log.Printf("Error sending response message: %v", sendingMessageError)
		}
	}
}
