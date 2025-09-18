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
	myWordListService     domain.Service
	log                   logger.Logger
}

type CallbackData struct {
	Action string `json:"action,omitempty"`
	Type   string `json:"request,omitempty"`
	Word   string `json:"word,omitempty"`
}

func NewCallbackHandler(
	irregularVerbsService domain.Service,
	myWordListService domain.Service,
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

	if strings.Contains(callbackData, "MyWordList") {
		if err := h.myWordListService.HandleCallback(ctx, update.CallbackQuery, bot); err != nil {
			h.log.Error(ctx, "Failed to handle word list callback", err,
				logger.Field{Key: "callback_data", Value: callbackData},
			)
			return h.answerCallback(ctx, bot, update.CallbackQuery.ID, "Ошибка обработки")
		}
		return h.answerCallback(ctx, bot, update.CallbackQuery.ID, "")
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

	// Here you would implement the actual save logic
	h.log.Info(ctx, "Word save requested",
		logger.Field{Key: "word", Value: word},
		logger.Field{Key: "user_id", Value: query.From.ID},
	)

	return h.answerCallback(ctx, bot, query.ID, fmt.Sprintf("✅ Слово '%s' добавлено в список!", word))
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