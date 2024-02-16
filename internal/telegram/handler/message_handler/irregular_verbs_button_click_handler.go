package message_handler

import (
	"Yulia-Lingo/internal/telegram/handler/callback_handler"
	utilService "Yulia-Lingo/internal/util_services"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func HandleIrregularVerbsButtonClick(bot *tgbotapi.BotAPI, chatID int64) error {
	inlineKeyboardMarkup, err := CreateLetterKeyboardMarkup()
	if err != nil {
		return fmt.Errorf("failed to create inlineKeyboardMarkup: %v", err)
	}

	messageText := "*С какой буквы вы хотите начать изучение неправильных глаголов?*\n\n"
	messageToUser := tgbotapi.NewMessage(chatID, messageText)
	messageToUser.ParseMode = "Markdown"
	messageToUser.ReplyMarkup = inlineKeyboardMarkup

	_, err = bot.Send(&messageToUser)
	if err != nil {
		return fmt.Errorf("failed to send the message for 'IrregularVerbs' button to a user: %v", err)
	}
	return nil
}

func CreateLetterKeyboardMarkup() (*tgbotapi.InlineKeyboardMarkup, error) {
	var rows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for _, letter := range letters {
		letterAsString := string(letter)
		requestData := callback_handler.KeyboardVerbValue{
			Request: "IrregularVerbs",
			Page:    0,
			Latter:  letterAsString,
		}
		jsonAsString, err := utilService.ConvertToJson(requestData)
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
