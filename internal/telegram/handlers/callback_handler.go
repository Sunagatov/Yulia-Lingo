package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/domain"
	"Yulia-Lingo/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	maxCallbackDataLength = 64
)

type CallbackHandler struct {
	irregularVerbsService domain.Service
	myWordListService     domain.WordListService
	log                   logger.Logger
}

type CallbackData struct {
	Action       string `json:"action,omitempty"`
	Type         string `json:"request,omitempty"`
	Word         string `json:"word,omitempty"`
	PartOfSpeech string `json:"part_of_speech,omitempty"`
}

func NewCallbackHandler(
	irregularVerbsService domain.Service,
	myWordListService domain.WordListService,
	log logger.Logger,
) *CallbackHandler {
	return &CallbackHandler{
		irregularVerbsService: irregularVerbsService,
		myWordListService:     myWordListService,
		log:                   log,
	}
}

func (h *CallbackHandler) Handle(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	if bot == nil {
		return fmt.Errorf("bot instance is nil")
	}

	if update.CallbackQuery == nil {
		return fmt.Errorf("callback query is nil")
	}

	callbackData := strings.TrimSpace(update.CallbackQuery.Data)

	h.log.Debug(ctx, "Processing callback",
		logger.Field{Key: "user_id", Value: update.CallbackQuery.From.ID},
		logger.Field{Key: "callback_data_length", Value: len(callbackData)},
	)

	// Validate callback data length
	if len(callbackData) > maxCallbackDataLength {
		h.log.Warn(ctx, "Callback data too long",
			logger.Field{Key: "length", Value: len(callbackData)},
			logger.Field{Key: "user_id", Value: update.CallbackQuery.From.ID},
		)
		return h.answerCallback(ctx, bot, update.CallbackQuery.ID, "Ошибка: некорректные данные")
	}

	// Check for back button first
	if callbackData == "b" {
		return h.handleWordListBack(ctx, bot, update.CallbackQuery)
	}

	// Check if it's irregular verbs or word list JSON
	if strings.Contains(callbackData, "IrregularVerbs") {
		if err := h.irregularVerbsService.HandleCallback(ctx, update.CallbackQuery, bot); err != nil {
			h.log.Error(ctx, "Failed to handle irregular verbs callback", err,
				logger.Field{Key: "callback_data", Value: callbackData},
			)
			return h.answerCallback(ctx, bot, update.CallbackQuery.ID, "Ошибка обработки")
		}
		return h.answerCallback(ctx, bot, update.CallbackQuery.ID, "")
	}

	// Handle word list callbacks (compact format: l:pos:page, r:id:pos:page, p:pos:page, n:pos:page)
	if strings.HasPrefix(callbackData, "l:") || strings.HasPrefix(callbackData, "r:") || strings.HasPrefix(callbackData, "p:") || strings.HasPrefix(callbackData, "n:") || strings.Contains(callbackData, "MyWordList") {
		if err := h.myWordListService.HandleCallback(ctx, update.CallbackQuery, bot); err != nil {
			h.log.Error(ctx, "Failed to handle word list callback", err,
				logger.Field{Key: "callback_data", Value: callbackData},
			)
			return h.answerCallback(ctx, bot, update.CallbackQuery.ID, "Ошибка обработки")
		}
		return h.answerCallback(ctx, bot, update.CallbackQuery.ID, "")
	}

	// Handle save word callback
	if strings.HasPrefix(callbackData, "s:") {
		return h.handleSaveWordCallback(ctx, bot, update.CallbackQuery, callbackData)
	}

	// Try to parse as translation action JSON
	var parsedData CallbackData
	if err := json.Unmarshal([]byte(callbackData), &parsedData); err == nil {
		return h.handleStructuredCallback(ctx, bot, update.CallbackQuery, &parsedData)
	}

	return h.answerCallback(ctx, bot, update.CallbackQuery.ID, "Неизвестная команда")
}

func (h *CallbackHandler) handleStructuredCallback(ctx context.Context, bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, data *CallbackData) error {
	switch data.Action {
	case "save_word":
		return h.handleSaveWord(ctx, bot, query, data.Word)
	case "save_word_final":
		return h.handleSaveWordFinal(ctx, bot, query, data)
	case "mark_learned":
		return h.handleMarkLearned(ctx, bot, query, data.Word)
	default:
		return h.answerCallback(ctx, bot, query.ID, "Неизвестное действие")
	}
}

func (h *CallbackHandler) handleSaveWord(ctx context.Context, bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, word string) error {
	if word == "" {
		return h.answerCallback(ctx, bot, query.ID, "Ошибка: пустое слово")
	}

	// Show part of speech selection for saving
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

	// Get translation from the original message
	translation := h.extractTranslationFromMessage(query.Message.Text)
	if translation == "" {
		translation = "Перевод не найден"
	}

	// Save word using the word list service
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

func (h *CallbackHandler) extractTranslationFromMessage(messageText string) string {
	// Simple extraction - get the first translation line
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

	// Here you would implement the actual mark as learned logic
	h.log.Info(ctx, "Word marked as learned",
		logger.Field{Key: "word", Value: word},
		logger.Field{Key: "user_id", Value: query.From.ID},
	)

	return h.answerCallback(ctx, bot, query.ID, fmt.Sprintf("🎉 Слово '%s' отмечено как выученное!", word))
}

func (h *CallbackHandler) answerCallback(ctx context.Context, bot *tgbotapi.BotAPI, callbackQueryID, text string) error {
	callback := tgbotapi.NewCallback(callbackQueryID, text)
	if _, err := bot.Request(callback); err != nil {
		h.log.Error(ctx, "Failed to answer callback query", err,
			logger.Field{Key: "callback_id", Value: callbackQueryID},
		)
		return fmt.Errorf("failed to answer callback query: %w", err)
	}
	return nil
}
