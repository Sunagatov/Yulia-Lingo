package domain

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Service defines common service operations
type Service interface {
	HandleButtonClick(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64) error
	HandleCallback(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error
}

// WordListService defines word list specific operations
type WordListService interface {
	Service
	AddWord(ctx context.Context, userID int64, word, partOfSpeech, translation string) error
}

// Repository defines common repository operations
type Repository interface {
	Initialize(ctx context.Context) error
}
