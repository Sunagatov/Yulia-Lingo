package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/logger"
	"Yulia-Lingo/internal/util"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	maxTranslations     = 5
	maxMessageLength    = 4096
	saveWordCallback    = "save_word"
	markLearnedCallback = "mark_learned"
)

type Service struct {
	apiClient APIClient
	log       logger.Logger
}

func NewService(apiClient APIClient, log logger.Logger) *Service {
	return &Service{
		apiClient: apiClient,
		log:       log,
	}
}

func (s *Service) HandleMessage(ctx context.Context, bot *tgbotapi.BotAPI, text string, chatID int64) error {
	if bot == nil {
		return fmt.Errorf("bot instance is nil")
	}

	if !util.IsValidEnglishWord(text) {
		return s.sendInvalidWordMessage(ctx, bot, chatID)
	}

	translation, err := s.apiClient.Translate(ctx, text)
	if err != nil {
		s.log.Error(ctx, "Failed to translate word", err,
			logger.Field{Key: "word", Value: text},
			logger.Field{Key: "chat_id", Value: chatID},
		)
		return s.sendErrorMessage(ctx, bot, chatID)
	}

	if err := translation.Validate(); err != nil {
		s.log.Warn(ctx, "Invalid translation received",
			logger.Field{Key: "word", Value: text},
			logger.Field{Key: "error", Value: err.Error()},
		)
		return s.sendErrorMessage(ctx, bot, chatID)
	}

	formattedTranslation := s.formatTranslation(text, translation)
	return s.sendTranslationMessage(ctx, bot, chatID, text, formattedTranslation)
}

func (s *Service) sendInvalidWordMessage(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64) error {
	messageText := "❌ *Некорректное слово*\n\n" +
		"Пожалуйста, отправьте корректное слово на английском языке.\n" +
		"Слово должно содержать только буквы, дефисы и апострофы."

	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		s.log.Error(ctx, "Failed to send invalid word message", err,
			logger.Field{Key: "chat_id", Value: chatID},
		)
		return fmt.Errorf("failed to send invalid word message: %w", err)
	}

	return nil
}

func (s *Service) sendErrorMessage(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64) error {
	messageText := "⚠️ *Ошибка перевода*\n\n" +
		"К сожалению, не удалось получить перевод слова.\n" +
		"Попробуйте еще раз позже."

	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ParseMode = "Markdown"

	if _, err := bot.Send(msg); err != nil {
		s.log.Error(ctx, "Failed to send error message", err,
			logger.Field{Key: "chat_id", Value: chatID},
		)
		return fmt.Errorf("failed to send error message: %w", err)
	}

	return nil
}

func (s *Service) sendTranslationMessage(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64, word, text string) error {
	if len(text) > maxMessageLength {
		text = text[:maxMessageLength-3] + "..."
	}

	keyboard := s.createActionKeyboard(word)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	if _, err := bot.Send(msg); err != nil {
		s.log.Error(ctx, "Failed to send translation message", err,
			logger.Field{Key: "chat_id", Value: chatID},
			logger.Field{Key: "word", Value: word},
		)
		return fmt.Errorf("failed to send translation message: %w", err)
	}

	s.log.Debug(ctx, "Sent translation message",
		logger.Field{Key: "chat_id", Value: chatID},
		logger.Field{Key: "word", Value: word},
	)

	return nil
}

func (s *Service) createActionKeyboard(word string) *tgbotapi.InlineKeyboardMarkup {
	saveData := map[string]string{
		"action": saveWordCallback,
		"word":   word,
	}
	learnedData := map[string]string{
		"action": markLearnedCallback,
		"word":   word,
	}

	saveJSON, _ := json.Marshal(saveData)
	learnedJSON, _ := json.Marshal(learnedData)

	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("💾 Добавить в список для изучения", string(saveJSON)),
			},
			{
				tgbotapi.NewInlineKeyboardButtonData("✅ Пометить как выученное", string(learnedJSON)),
			},
		},
	}
}

func (s *Service) formatTranslation(word string, translation Translation) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("🔤 *Перевод слова:* `%s`\n", word))
	builder.WriteString(util.GetMessageDelimiter() + "\n\n")

	for i, entry := range translation.Dictionary {
		if i >= maxTranslations {
			break
		}

		entry.Sanitize()
		builder.WriteString(fmt.Sprintf("📝 *%s*\n", entry.PartOfSpeech))

		if len(entry.Terms) > 0 {
			maxTerms := maxTranslations
			if maxTerms > len(entry.Terms) {
				maxTerms = len(entry.Terms)
			}

			terms := entry.Terms[:maxTerms]
			for j, term := range terms {
				builder.WriteString(fmt.Sprintf("%d. %s\n", j+1, term))
			}
		}

		if i < len(translation.Dictionary)-1 && i < maxTranslations-1 {
			builder.WriteString("\n")
		}
	}

	builder.WriteString("\n" + util.GetMessageDelimiter())
	builder.WriteString("\n\n💡 *Выберите действие ниже:*")

	return builder.String()
}