package handler

import (
	"Yulia-Lingo/internal/api"
	tgbutton "Yulia-Lingo/internal/telegram/button"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"regexp"
)

func HandleMessageFromUser(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update) {
	messageFromUser := botUpdate.Message
	chatID := messageFromUser.Chat.ID

	textFromUser := messageFromUser.Text

	switch textFromUser {
	case tgbutton.StartButtonName:
		tgbutton.HandleStartButtonClick(bot, chatID)
	case tgbutton.IrregularVerbListButtonName:
		tgbutton.HandleIrregularVerbsListButtonClick(bot, chatID)
	default:
		handleDefaultCaseUserMessage(bot, textFromUser, chatID)
	}
}

func handleDefaultCaseUserMessage(bot *tgbotapi.BotAPI, textFromUser string, chatID int64) {
	if !isValidWord(textFromUser) {
		responseMessage := "Пожалуйста, отправьте корректное слово на английском языке"
		messageToUser := tgbotapi.NewMessage(chatID, responseMessage)
		_, sendingMessageError := bot.Send(messageToUser)
		if sendingMessageError != nil {
			log.Printf("Error sending message to a user: %v", sendingMessageError)
		}
	}

	responseMessage, err := api.RequestTranslateAPI(textFromUser)
	if err != nil {
		log.Printf("Error fetching from API: %v", err)
		responseMessage = "Sorry, there was an error processing your request."
	}
	messageToUser := tgbotapi.NewMessage(chatID, responseMessage)

	messageToUser.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Сохранить", "save_word_option"),
		),
	)
	_, errorMessage := bot.Send(messageToUser)
	if errorMessage != nil {
		log.Printf("Error sending response message: %v", errorMessage)
	}
}

func isValidWord(word string) bool {
	const pattern = `^[A-Za-z]+$`
	matched, _ := regexp.MatchString(pattern, word)
	return matched
}
