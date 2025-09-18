package irregular_verbs

import (
	"context"
	"encoding/json"
	"fmt"

	"Yulia-Lingo/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	verbsPerPage     = 10
	maxMessageLength = 4096
	requestType      = "IrregularVerbs"
)

type Service struct {
	repository Repository
	log        logger.Logger
}

func NewService(repository Repository, log logger.Logger) *Service {
	return &Service{
		repository: repository,
		log:        log,
	}
}

func (s *Service) HandleButtonClick(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64) error {
	if bot == nil {
		return fmt.Errorf("bot instance is nil")
	}

	inlineKeyboard, err := s.createLetterKeyboard(ctx)
	if err != nil {
		return fmt.Errorf("failed to create keyboard: %w", err)
	}

	messageText := "С какой буквы вы хотите начать изучение неправильных глаголов?\n\n" +
		"Выберите первую букву глагола из списка ниже:"

	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ReplyMarkup = inlineKeyboard

	if _, err = bot.Send(msg); err != nil {
		s.log.Error(ctx, "Failed to send irregular verbs message", err,
			logger.Field{Key: "chat_id", Value: chatID},
		)
		return fmt.Errorf("failed to send irregular verbs message: %w", err)
	}

	s.log.Debug(ctx, "Sent irregular verbs letter selection",
		logger.Field{Key: "chat_id", Value: chatID},
	)

	return nil
}

func (s *Service) HandleCallback(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error {
	if callbackQuery == nil {
		return fmt.Errorf("callback query is nil")
	}

	var keyboardValue KeyboardVerbValue
	if err := json.Unmarshal([]byte(callbackQuery.Data), &keyboardValue); err != nil {
		s.log.Error(ctx, "Failed to unmarshal callback data", err,
			logger.Field{Key: "data", Value: callbackQuery.Data},
		)
		return fmt.Errorf("invalid callback data: %w", err)
	}

	if err := keyboardValue.Validate(); err != nil {
		return fmt.Errorf("invalid keyboard value: %w", err)
	}

	// Check if this is a "back to letters" request
	if keyboardValue.Letter == "BACK_TO_LETTERS" {
		return s.showLetterSelection(ctx, callbackQuery, bot)
	}

	return s.handleVerbListCallback(ctx, callbackQuery, bot, &keyboardValue)
}

func (s *Service) handleVerbListCallback(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, keyboardValue *KeyboardVerbValue) error {
	offset := keyboardValue.Page * verbsPerPage
	verbs, err := s.repository.GetPage(ctx, offset, verbsPerPage, keyboardValue.Letter)
	if err != nil {
		return fmt.Errorf("failed to get verbs page: %w", err)
	}

	totalCount, err := s.repository.GetTotalCount(ctx, keyboardValue.Letter)
	if err != nil {
		return fmt.Errorf("failed to get total count: %w", err)
	}

	messageText := s.formatVerbsList(keyboardValue.Letter, verbs, keyboardValue.Page, totalCount)
	keyboard := s.createNavigationKeyboard(keyboardValue, totalCount)

	editMsg := tgbotapi.NewEditMessageText(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		messageText,
	)
	editMsg.ParseMode = "Markdown"
	editMsg.ReplyMarkup = &keyboard

	if _, err := bot.Send(editMsg); err != nil {
		s.log.Error(ctx, "Failed to edit message", err,
			logger.Field{Key: "chat_id", Value: callbackQuery.Message.Chat.ID},
			logger.Field{Key: "message_id", Value: callbackQuery.Message.MessageID},
		)
		return fmt.Errorf("failed to edit message: %w", err)
	}

	return nil
}

func (s *Service) showLetterSelection(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error {
	inlineKeyboard, err := s.createLetterKeyboard(ctx)
	if err != nil {
		return fmt.Errorf("failed to create keyboard: %w", err)
	}

	messageText := "С какой буквы вы хотите начать изучение неправильных глаголов?\n\n" +
		"Выберите первую букву глагола из списка ниже:"

	editMsg := tgbotapi.NewEditMessageText(
		callbackQuery.Message.Chat.ID,
		callbackQuery.Message.MessageID,
		messageText,
	)
	editMsg.ReplyMarkup = inlineKeyboard

	if _, err := bot.Send(editMsg); err != nil {
		s.log.Error(ctx, "Failed to edit message for letter selection", err,
			logger.Field{Key: "chat_id", Value: callbackQuery.Message.Chat.ID},
			logger.Field{Key: "message_id", Value: callbackQuery.Message.MessageID},
		)
		return fmt.Errorf("failed to edit message: %w", err)
	}

	return nil
}
