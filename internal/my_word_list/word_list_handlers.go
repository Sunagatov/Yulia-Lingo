package my_word_list

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (s *Service) handleRemoveWordCompact(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, callbackData string, userID int64) error {
	parts := strings.Split(callbackData, ":")
	if len(parts) != 4 {
		return fmt.Errorf("invalid remove callback data")
	}

	wordID, _ := strconv.Atoi(parts[1])
	partOfSpeech := parts[2]
	page, _ := strconv.Atoi(parts[3])

	if err := s.repository.RemoveWord(ctx, userID, wordID); err != nil {
		s.log.Error(ctx, "Failed to remove word", err)
		return fmt.Errorf("failed to remove word: %w", err)
	}

	keyboardValue := &KeyboardWordValue{
		Request:      "MyWordList",
		Page:         page,
		PartOfSpeech: partOfSpeech,
	}
	return s.showWordList(ctx, callbackQuery, bot, keyboardValue, userID)
}

func (s *Service) handlePageNavigationCompact(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, callbackData string, userID int64) error {
	parts := strings.Split(callbackData, ":")
	if len(parts) != 3 {
		return fmt.Errorf("invalid navigation callback data")
	}

	partOfSpeech := parts[1]
	page, _ := strconv.Atoi(parts[2])

	keyboardValue := &KeyboardWordValue{
		Request:      "MyWordList",
		Page:         page,
		PartOfSpeech: partOfSpeech,
	}
	return s.showWordList(ctx, callbackQuery, bot, keyboardValue, userID)
}

func (s *Service) handleShowWordListCompact(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, callbackData string, userID int64) error {
	parts := strings.Split(callbackData, ":")
	if len(parts) != 3 {
		return fmt.Errorf("invalid list callback data")
	}

	partOfSpeech := parts[1]
	page, _ := strconv.Atoi(parts[2])

	keyboardValue := &KeyboardWordValue{
		Request:      "MyWordList",
		Page:         page,
		PartOfSpeech: partOfSpeech,
	}
	return s.showWordList(ctx, callbackQuery, bot, keyboardValue, userID)
}
