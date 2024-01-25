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
		processUpdate(telegramBot, telegramBotUpdate)
	}
}

func processUpdate(telegramBot *tgbotapi.BotAPI, update tgbotapi.Update) {
	messageFromUser := update.Message.Text
	chatID := update.Message.Chat.ID

	if !isValidWord(messageFromUser) {
		sendErrorMessage(telegramBot, chatID, "Please send a single, valid word in English.")
		return
	}

	responseMessage, err := api.RequestWordsAPI(messageFromUser)
	if err != nil {
		log.Printf("Error fetching from API: %v", err)
		responseMessage = "Sorry, there was an error processing your request."
	}
	sendMessageToUser(telegramBot, chatID, responseMessage)
}

func isValidWord(word string) bool {
	const pattern = `^[A-Za-z]+$`
	matched, _ := regexp.MatchString(pattern, word)
	return matched
}

func sendErrorMessage(telegramBot *tgbotapi.BotAPI, chatID int64, message string) {
	messageToUser := tgbotapi.NewMessage(chatID, message)
	_, err := telegramBot.Send(messageToUser)
	if err != nil {
		log.Printf("Error sending message to a user: %v", err)
	}
}

func sendMessageToUser(telegramBot *tgbotapi.BotAPI, chatID int64, message string) {
	messageToUser := tgbotapi.NewMessage(chatID, message)
	_, err := telegramBot.Send(messageToUser)
	if err != nil {
		log.Printf("Error sending message to a user: %v", err)
	}
}
