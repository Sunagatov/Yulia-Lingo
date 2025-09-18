package my_word_list

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"Yulia-Lingo/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	keyboard := s.createPartsOfSpeechKeyboard()

	messageText := "*📚 Мой список слов*\n\n" +
		"Выберите часть речи для просмотра ваших сохраненных слов:\n\n" +
		"💡 *Совет:* Отправьте любое слово для перевода и добавления в список!"
	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	_, err := bot.Send(msg)
	if err != nil {
		s.log.Error(ctx, "Failed to send word list message", err,
			logger.Field{Key: "chat_id", Value: chatID},
		)
		return fmt.Errorf("failed to send word list message: %w", err)
	}

	return nil
}

func (s *Service) HandleCallback(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error {
	callbackData := callbackQuery.Data
	userID := callbackQuery.From.ID

	// Handle compact format callbacks
	if strings.HasPrefix(callbackData, "l:") {
		return s.handleShowWordListCompact(ctx, callbackQuery, bot, callbackData, userID)
	}
	if strings.HasPrefix(callbackData, "r:") {
		return s.handleRemoveWordCompact(ctx, callbackQuery, bot, callbackData, userID)
	}
	if strings.HasPrefix(callbackData, "p:") || strings.HasPrefix(callbackData, "n:") {
		return s.handlePageNavigationCompact(ctx, callbackQuery, bot, callbackData, userID)
	}

	// Handle JSON format callbacks
	var keyboardValue KeyboardWordValue
	if err := json.Unmarshal([]byte(callbackData), &keyboardValue); err != nil {
		s.log.Error(ctx, "Failed to unmarshal callback data", err,
			logger.Field{Key: "data", Value: callbackData},
		)
		return fmt.Errorf("invalid callback data: %w", err)
	}

	return s.showWordList(ctx, callbackQuery, bot, &keyboardValue, userID)
}

func (s *Service) AddWord(ctx context.Context, userID int64, word, partOfSpeech, translation string) error {
	return s.repository.AddWord(ctx, userID, word, partOfSpeech, translation)
}

func (s *Service) showWordList(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, keyboardValue *KeyboardWordValue, userID int64) error {
	const pageSize = 5
	offset := keyboardValue.Page * pageSize

	totalCount, err := s.repository.GetTotalCount(ctx, userID, keyboardValue.PartOfSpeech)
	if err != nil {
		s.log.Error(ctx, "Failed to get total count", err)
		return fmt.Errorf("failed to get total count: %w", err)
	}

	if totalCount == 0 {
		messageText := fmt.Sprintf("*📚 %s*\n\nУ вас пока нет сохраненных слов в этой категории.\n\n💡 Отправьте любое слово для перевода и добавления!", s.getPartOfSpeechName(keyboardValue.PartOfSpeech))
		return s.editMessage(ctx, callbackQuery, bot, messageText, s.createBackKeyboard())
	}

	words, err := s.repository.GetPage(ctx, userID, offset, pageSize, keyboardValue.PartOfSpeech)
	if err != nil {
		s.log.Error(ctx, "Failed to get words page", err)
		return fmt.Errorf("failed to get words page: %w", err)
	}

	messageText := s.formatWordList(keyboardValue.PartOfSpeech, words, keyboardValue.Page+1, (totalCount+pageSize-1)/pageSize)
	keyboard := s.createWordListKeyboard(words, keyboardValue, totalCount, pageSize)

	return s.editMessage(ctx, callbackQuery, bot, messageText, keyboard)
}

func (s *Service) editMessage(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, text string, keyboard *tgbotapi.InlineKeyboardMarkup) error {
	editMsg := tgbotapi.NewEditMessageText(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, text)
	editMsg.ParseMode = "Markdown"
	editMsg.ReplyMarkup = keyboard

	if _, err := bot.Send(editMsg); err != nil {
		s.log.Error(ctx, "Failed to edit message", err)
		return fmt.Errorf("failed to edit message: %w", err)
	}

	callback := tgbotapi.NewCallback(callbackQuery.ID, "")
	_, err := bot.Request(callback)
	return err
}
