package my_word_list

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (s *Service) createWordListKeyboard(words []Entity, keyboardValue *KeyboardWordValue, totalCount, pageSize int) *tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Remove buttons for each word
	for _, word := range words {
		callbackData := fmt.Sprintf("r:%d:%s:%d", word.ID, keyboardValue.PartOfSpeech, keyboardValue.Page)
		btn := tgbotapi.NewInlineKeyboardButtonData("❌ "+word.Word, callbackData)
		rows = append(rows, []tgbotapi.InlineKeyboardButton{btn})
	}

	// Navigation buttons
	var navRow []tgbotapi.InlineKeyboardButton
	if keyboardValue.Page > 0 {
		callbackData := fmt.Sprintf("p:%s:%d", keyboardValue.PartOfSpeech, keyboardValue.Page-1)
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", callbackData))
	}
	if (keyboardValue.Page+1)*pageSize < totalCount {
		callbackData := fmt.Sprintf("n:%s:%d", keyboardValue.PartOfSpeech, keyboardValue.Page+1)
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("Далее ➡️", callbackData))
	}
	if len(navRow) > 0 {
		rows = append(rows, navRow)
	}

	// Back button
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🔙 Назад к категориям", "b"),
	})

	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func (s *Service) createBackKeyboard() *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{{
			tgbotapi.NewInlineKeyboardButtonData("🔙 Назад к категориям", "b"),
		}},
	}
}

func (s *Service) createPartsOfSpeechKeyboard() *tgbotapi.InlineKeyboardMarkup {
	partsOfSpeech := []struct {
		key  string
		name string
	}{
		{"adjective", "Прилагательное"},
		{"adverb", "Наречие"},
		{"noun", "Существительное"},
		{"preposition", "Предлог"},
		{"pronoun", "Местоимение"},
		{"verb", "Глагол"},
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for _, pos := range partsOfSpeech {
		callbackData := fmt.Sprintf("l:%s:0", pos.key)
		btn := tgbotapi.NewInlineKeyboardButtonData(pos.name, callbackData)
		currentRow = append(currentRow, btn)

		if len(currentRow) == 2 {
			rows = append(rows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}

	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}
