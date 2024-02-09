package button

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func CreateLetterKeyboardMarkup() tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for _, letter := range letters {
		btn := tgbotapi.NewInlineKeyboardButtonData(string(letter), LetterSelectionPrefix+string(letter))
		currentRow = append(currentRow, btn)

		if len(currentRow) == 5 {
			rows = append(rows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
