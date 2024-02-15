package handler

import (
	irregularVerbsManager "Yulia-Lingo/internal/database/irregular_verbs"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

const (
	letters            = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	greetingBotMessage = "Здравствуйте, %s %s!\n\nЭто телеграм бот - Yulia-lingo.\n\n" +
		"Бот поможет вам пополнить словарный запас английского языка.\n\n" +
		"Сейчас доступен:\n- Список неправильных глаголов."
)

func HandleMessageFromUser(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update) error {
	chatID := botUpdate.Message.Chat.ID
	messageFromUser := botUpdate.Message.Text

	switch messageFromUser {
	case "/start":
		{
			userFirstName := botUpdate.Message.From.FirstName
			userLastName := botUpdate.Message.From.LastName
			greetingMessage := fmt.Sprintf(greetingBotMessage, userFirstName, userLastName)
			messageToUser := tgbotapi.NewMessage(chatID, greetingMessage)
			messageToUser.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("🔺 Неправильные глаголы")),
			)
			_, errorMessage := bot.Send(&messageToUser)
			if errorMessage != nil {
				return fmt.Errorf("failed to send the greeting message to a user: %v", errorMessage)
			}
		}
	case "🔺 Неправильные глаголы":
		{
			inlineKeyboardMarkup, err := CreateLetterKeyboardMarkup()
			if err != nil {
				return fmt.Errorf("failed to create inlineKeyboardMarkup: %v", err)
			}
			log.Printf("inlineKeyboardMarkup: %v", inlineKeyboardMarkup)

			messageText := "С какой буквы вы хотите начать изучение неправильных глаголов?\n\n"
			messageToUser := tgbotapi.NewMessage(chatID, messageText)
			messageToUser.ReplyMarkup = inlineKeyboardMarkup

			_, err = bot.Send(&messageToUser)
			if err != nil {
				return fmt.Errorf("failed to send the message for 'IrregularVerbs' button to a user: %v", err)
			}
		}
	default:
	}
	return nil
}

func CreateLetterKeyboardMarkup() (*tgbotapi.InlineKeyboardMarkup, error) {
	var rows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for _, letter := range letters {
		letterAsString := string(letter)
		requestData := irregularVerbsManager.KeyboardVerbValue{
			Request: "IrregularVerbs",
			Page:    0,
			Latter:  letterAsString,
		}
		jsonAsString, err := irregularVerbsManager.ConvertToJson(requestData)
		if err != nil {
			return nil, fmt.Errorf("failed to create JSON for letter '%v': %v", letterAsString, err)
		}
		btn := tgbotapi.NewInlineKeyboardButtonData(letterAsString, jsonAsString)
		currentRow = append(currentRow, btn)

		if len(currentRow) == 5 {
			rows = append(rows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}, nil
}
