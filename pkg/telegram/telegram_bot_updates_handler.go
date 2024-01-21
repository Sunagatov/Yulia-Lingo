package telegram

import (
	"Yulia-Lingo/internal/word/api"
	"log"
	"regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func LaunchTelegramBot(telegramBot *tgbotapi.BotAPI) {
	updateEndpoint := "/" + telegramBot.Token
	telegramBotUpdates := telegramBot.ListenForWebhook(updateEndpoint)

	for telegramBotUpdate := range telegramBotUpdates {
		if telegramBotUpdate.Message == nil {
			continue
		}
		messageFromUser := telegramBotUpdate.Message.Text
		chatID := telegramBotUpdate.Message.Chat.ID

		if !isValidWord(messageFromUser) {
			responseMessage := "Please send a single, valid word in English."
			messageToUser := tgbotapi.NewMessage(chatID, responseMessage)
			_, sendingMessageError := telegramBot.Send(messageToUser)
			if sendingMessageError != nil {
				log.Printf("Error sending message to a user: %v", sendingMessageError)
			}
			continue
		}
		responseMessage, err := api.RequestWordsAPI(messageFromUser)
		if err != nil {
			log.Printf("Error fetching from API: %v", err)
			responseMessage = "Sorry, there was an error processing your request."
		}
		messageToUser := tgbotapi.NewMessage(chatID, responseMessage)

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
