package handler

import (
	"Yulia-Lingo/internal/api"
	tgbutton "Yulia-Lingo/internal/telegram/button"
	"Yulia-Lingo/internal/telegram/message"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"regexp"
)

func HandleMessageFromUser(botUpdate tgbotapi.Update) {
	messageFromUser := botUpdate.Message
	chatID := messageFromUser.Chat.ID
	textFromUser := messageFromUser.Text

	switch textFromUser {
	case tgbutton.StartButtonName:
		tgbutton.HandleStartButtonClick(chatID)
	case tgbutton.IrregularVerbListButtonName:
		tgbutton.HandleIrregularVerbsListButtonClick(chatID)
	default:
		handleDefaultCaseUserMessage(textFromUser, chatID)
	}
}

func handleDefaultCaseUserMessage(textFromUser string, chatID int64) {
	if err := inputWordValidator(textFromUser, chatID); err != nil {
		return
	}

	messageToUser := translateInputWords(textFromUser, chatID)
	addKeyboardMarkup(&messageToUser)

	errorMessage := message.Send(&messageToUser)
	if errorMessage != nil {
		log.Printf("Error sending response message: %v", errorMessage)
	}
}

func inputWordValidator(textFromUser string, chatID int64) error {
	if !isValidWord(textFromUser) {
		responseMessage := "Пожалуйста, отправьте корректное слово на английском языке"
		messageToUser := tgbotapi.NewMessage(chatID, responseMessage)
		sendingMessageError := message.Send(&messageToUser)
		if sendingMessageError != nil {
			return fmt.Errorf("error sending message to a user: %v", sendingMessageError)
		}
		return fmt.Errorf("incorrect format of text")
	}

	return nil
}

func translateInputWords(textFromUser string, chatID int64) tgbotapi.MessageConfig {
	responseMessage, err := api.RequestTranslateAPI(textFromUser)
	if err != nil {
		log.Printf("Error fetching from API: %v", err)
		responseMessage = "Sorry, there was an error processing your request."
	}
	return tgbotapi.NewMessage(chatID, responseMessage)
}

func addKeyboardMarkup(messageToUser *tgbotapi.MessageConfig) {
	messageToUser.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Сохранить", "save_word_option"),
		),
	)
}

func isValidWord(word string) bool {
	const pattern = `^[A-Za-z]+$`
	matched, _ := regexp.MatchString(pattern, word)
	return matched
}
