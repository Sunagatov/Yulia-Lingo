package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *CallbackHandler) extractTranslationFromMessage(messageText string) string {
	lines := strings.Split(messageText, "\n")
	for _, line := range lines {
		if strings.Contains(line, "1. ") {
			return strings.TrimPrefix(line, "1. ")
		}
	}
	return ""
}

func (h *CallbackHandler) getPartOfSpeechName(partOfSpeech string) string {
	names := map[string]string{
		"noun":        "сущ.",
		"verb":        "гл.",
		"adjective":   "прил.",
		"adverb":      "нар.",
		"preposition": "предл.",
		"pronoun":     "мест.",
	}
	if name, ok := names[partOfSpeech]; ok {
		return name
	}
	return partOfSpeech
}

func (h *CallbackHandler) handleWordListBack(ctx context.Context, bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery) error {
	messageText := "*📚 Мой список слов*\n\n" +
		"Выберите часть речи:\n\n" +
		"💡 Отправьте слово для перевода и сохранения!"

	keyboard := h.createPartsOfSpeechKeyboard()

	editMsg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, messageText)
	editMsg.ParseMode = "Markdown"
	editMsg.ReplyMarkup = keyboard

	if _, err := bot.Send(editMsg); err != nil {
		h.log.Error(ctx, "Failed to edit message for word list back", err)
		return fmt.Errorf("failed to edit message: %w", err)
	}

	return h.answerCallback(ctx, bot, query.ID, "")
}

func (h *CallbackHandler) createPartsOfSpeechKeyboard() *tgbotapi.InlineKeyboardMarkup {
	partsOfSpeech := map[string]string{
		"noun":        "Существительное",
		"verb":        "Глагол",
		"adjective":   "Прилагательное",
		"adverb":      "Наречие",
		"preposition": "Предлог",
		"pronoun":     "Местоимение",
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for _, abbreviation := range []string{"adjective", "adverb", "noun", "preposition", "pronoun", "verb"} {
		russianName := partsOfSpeech[abbreviation]

		requestData := map[string]interface{}{
			"request":        "MyWordList",
			"page":           0,
			"part_of_speech": abbreviation,
		}

		jsonData, _ := json.Marshal(requestData)
		btn := tgbotapi.NewInlineKeyboardButtonData(russianName, string(jsonData))
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
