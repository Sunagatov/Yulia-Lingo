package handlers

import (
	"context"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *CallbackHandler) handleSaveWord(ctx context.Context, bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string) error {
	if word == "" {
		return h.answerCallback(ctx, bot, query.ID, "Ошибка: пустое слово")
	}

	return h.showPartOfSpeechSelection(ctx, bot, query, word)
}

func (h *CallbackHandler) showPartOfSpeechSelection(ctx context.Context, bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string) error {
	messageText := fmt.Sprintf("📚 *Сохранение слова:* `%s`\n\nВыберите часть речи:", word)

	keyboard := h.createSaveWordKeyboard(word)

	editMsg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, messageText)
	editMsg.ParseMode = "Markdown"
	editMsg.ReplyMarkup = keyboard

	if _, err := bot.Send(editMsg); err != nil {
		h.log.Error(ctx, "Failed to edit message for word saving", err)
		return fmt.Errorf("failed to edit message: %w", err)
	}

	return h.answerCallback(ctx, bot, query.ID, "")
}

func (h *CallbackHandler) createSaveWordKeyboard(word string) *tgbotapi.InlineKeyboardMarkup {
	partsOfSpeech := []struct {
		key  string
		name string
	}{
		{"n", "Существительное"},
		{"v", "Глагол"},
		{"a", "Прилагательное"},
		{"d", "Наречие"},
		{"p", "Предлог"},
		{"r", "Местоимение"},
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, pos := range partsOfSpeech {
		callbackData := fmt.Sprintf("s:%s:%s", word, pos.key)
		btn := tgbotapi.NewInlineKeyboardButtonData(pos.name, callbackData)
		rows = append(rows, []tgbotapi.InlineKeyboardButton{btn})
	}

	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func (h *CallbackHandler) handleSaveWordFinal(ctx context.Context, bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data *CallbackData) error {
	if data.Word == "" || data.PartOfSpeech == "" {
		return h.answerCallback(ctx, bot, query.ID, "Ошибка: некорректные данные")
	}

	translation := h.extractTranslationFromMessage(query.Message.Text)
	if translation == "" {
		translation = "Перевод не найден"
	}

	if err := h.myWordListService.AddWord(ctx, query.From.ID, data.Word, data.PartOfSpeech, translation); err != nil {
		h.log.Error(ctx, "Failed to save word", err,
			logger.Field{Key: "word", Value: data.Word},
			logger.Field{Key: "user_id", Value: query.From.ID},
		)
		return h.answerCallback(ctx, bot, query.ID, "Ошибка сохранения")
	}

	messageText := fmt.Sprintf("✅ *Слово сохранено!*\n\n📝 **%s** (%s)\n🔄 %s\n\n📚 Посмотреть ваш список можно в меню 'Мой список слов'.", data.Word, h.getPartOfSpeechName(data.PartOfSpeech), translation)

	editMsg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, messageText)
	editMsg.ParseMode = "Markdown"

	if _, err := bot.Send(editMsg); err != nil {
		h.log.Error(ctx, "Failed to edit message after saving word", err)
		return fmt.Errorf("failed to edit message: %w", err)
	}

	return h.answerCallback(ctx, bot, query.ID, fmt.Sprintf("✅ Слово '%s' добавлено!", data.Word))
}

func (h *CallbackHandler) handleSaveWordCallback(ctx context.Context, bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, callbackData string) error {
	parts := strings.Split(callbackData, ":")
	if len(parts) != 3 {
		return h.answerCallback(ctx, bot, query.ID, "Ошибка данных")
	}

	word := parts[1]
	posKey := parts[2]

	posMap := map[string]string{
		"n": "noun",
		"v": "verb",
		"a": "adjective",
		"d": "adverb",
		"p": "preposition",
		"r": "pronoun",
	}

	partOfSpeech, ok := posMap[posKey]
	if !ok {
		return h.answerCallback(ctx, bot, query.ID, "Некорректная часть речи")
	}

	translation := h.extractTranslationFromMessage(query.Message.Text)
	if translation == "" {
		translation = "Перевод не найден"
	}

	if err := h.myWordListService.AddWord(ctx, query.From.ID, word, partOfSpeech, translation); err != nil {
		h.log.Error(ctx, "Failed to save word", err)
		return h.answerCallback(ctx, bot, query.ID, "Ошибка сохранения")
	}

	messageText := fmt.Sprintf("✅ *Слово сохранено!*\n\n📝 **%s** (%s)\n🔄 %s\n\n📚 Посмотреть ваш список можно в меню 'Мой список слов'.", word, h.getPartOfSpeechName(partOfSpeech), translation)

	editMsg := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, messageText)
	editMsg.ParseMode = "Markdown"

	if _, err := bot.Send(editMsg); err != nil {
		h.log.Error(ctx, "Failed to edit message after saving word", err)
		return fmt.Errorf("failed to edit message: %w", err)
	}

	return h.answerCallback(ctx, bot, query.ID, fmt.Sprintf("✅ Слово '%s' добавлено!", word))
}

func (h *CallbackHandler) handleMarkLearned(ctx context.Context, bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string) error {
	if word == "" {
		return h.answerCallback(ctx, bot, query.ID, "Ошибка: пустое слово")
	}

	h.log.Info(ctx, "Word marked as learned",
		logger.Field{Key: "word", Value: word},
		logger.Field{Key: "user_id", Value: query.From.ID},
	)

	return h.answerCallback(ctx, bot, query.ID, fmt.Sprintf("🎉 Слово '%s' отмечено как выученное!", word))
}
