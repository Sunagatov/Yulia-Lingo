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

	// Route to appropriate handler
	return h.routeCallback(ctx, bot, update.CallbackQuery, callbackData)
}

func (h *CallbackHandler) routeCallback(ctx context.Context, bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, callbackData string) error {
	// Check for back button first
	if callbackData == "b" {
		return h.handleWordListBack(ctx, bot, query)
	}

	// Check if it's irregular verbs
	if strings.Contains(callbackData, "IrregularVerbs") {
		if err := h.irregularVerbsService.HandleCallback(ctx, query, bot); err != nil {
			h.log.Error(ctx, "Failed to handle irregular verbs callback", err,
				logger.Field{Key: "callback_data", Value: callbackData},
			)
			return h.answerCallback(ctx, bot, query.ID, "Ошибка обработки")
		}
		return h.answerCallback(ctx, bot, query.ID, "")
	}

	// Handle word list callbacks (compact format)
	if h.isWordListCallback(callbackData) {
		if err := h.myWordListService.HandleCallback(ctx, query, bot); err != nil {
			h.log.Error(ctx, "Failed to handle word list callback", err,
				logger.Field{Key: "callback_data", Value: callbackData},
			)
			return h.answerCallback(ctx, bot, query.ID, "Ошибка обработки")
		}
		return h.answerCallback(ctx, bot, query.ID, "")
	}

	// Handle save word callback
	if strings.HasPrefix(callbackData, "s:") {
		return h.handleSaveWordCallback(ctx, bot, query, callbackData)
	}

	// Try to parse as translation action JSON
	var parsedData CallbackData
	if err := json.Unmarshal([]byte(callbackData), &parsedData); err == nil {
		return h.handleStructuredCallback(ctx, bot, query, &parsedData)
	}

	return h.answerCallback(ctx, bot, query.ID, "Неизвестная команда")
}

func (h *CallbackHandler) isWordListCallback(callbackData string) bool {
	return strings.HasPrefix(callbackData, "l:") ||
		strings.HasPrefix(callbackData, "r:") ||
		strings.HasPrefix(callbackData, "p:") ||
		strings.HasPrefix(callbackData, "n:") ||
		strings.Contains(callbackData, "MyWordList")
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
