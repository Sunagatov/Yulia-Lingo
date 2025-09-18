package domain

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MessageHandler defines the interface for handling Telegram messages
type MessageHandler interface {
	Handle(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error
}

// CallbackHandler defines the interface for handling Telegram callback queries
type CallbackHandler interface {
	Handle(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error
}

// Service defines common service operations
type Service interface {
	HandleButtonClick(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64) error
	HandleCallback(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error
}

// Repository defines common repository operations
type Repository interface {
	Initialize(ctx context.Context) error
}

// Translator defines translation service operations
type Translator interface {
	Translate(ctx context.Context, text string) (Translation, error)
}

// Translation represents a translation result
type Translation struct {
	Dictionary []DictionaryEntry `json:"dictionary"`
}

// DictionaryEntry represents a dictionary entry with part of speech and terms
type DictionaryEntry struct {
	PartOfSpeech string   `json:"part_of_speech"`
	Terms        []string `json:"terms"`
}