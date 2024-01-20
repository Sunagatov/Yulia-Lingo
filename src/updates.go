package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"regexp"
)

func listenForTelegramBotUpdates(telegramBot *tgbotapi.BotAPI) {
	updateEndpoint := "/" + telegramBot.Token
	telegramBotUpdates := telegramBot.ListenForWebhook(updateEndpoint)

	for telegramBotUpdate := range telegramBotUpdates {
		if telegramBotUpdate.Message == nil {
			continue
		}
		messageFromUser := telegramBotUpdate.Message.Text

		if !isValidWord(messageFromUser) {
			responseMessage := "Please send a single, valid word in English."
			messageToUser := tgbotapi.NewMessage(telegramBotUpdate.Message.Chat.ID, responseMessage)
			_, sendingMessageError := telegramBot.Send(messageToUser)
			if sendingMessageError != nil {
				log.Printf("Error sending message to a user: %v", sendingMessageError)
			}
			continue
		}
		responseMessage, err := requestWordsAPI(messageFromUser)
		if err != nil {
			log.Printf("Error fetching from API: %v", err)
			responseMessage = "Sorry, there was an error processing your request."
		}
		messageToUser := tgbotapi.NewMessage(telegramBotUpdate.Message.Chat.ID, responseMessage)

		_, errorMessage := telegramBot.Send(messageToUser)
		if errorMessage != nil {
			log.Printf("Error sending response message: %v", errorMessage)
		}
	}
}

func isValidWord(word string) bool {
	const pattern = `^[A-Za-z]+$`
	matched, _ := regexp.MatchString(pattern, word)
	return matched
}
