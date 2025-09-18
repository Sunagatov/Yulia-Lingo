package telegram

import (
	"fmt"

	"Yulia-Lingo/internal/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotManager struct {
	bot *tgbotapi.BotAPI
	cfg *config.Config
}

func NewBotManager(cfg *config.Config) (*BotManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if cfg.Telegram.BotToken == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	return &BotManager{
		bot: bot,
		cfg: cfg,
	}, nil
}

func (bm *BotManager) GetBot() *tgbotapi.BotAPI {
	return bm.bot
}

