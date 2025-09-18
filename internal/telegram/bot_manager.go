package telegram

import (
	"fmt"

	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
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

func (bm *BotManager) ConfigureWebhook() error {
	if bm.bot == nil {
		return fmt.Errorf("bot instance is nil")
	}

	if bm.cfg.Telegram.WebhookURL == "" {
		logger.Info("No webhook URL configured, skipping webhook setup")
		return nil
	}

	webhookConfig, err := tgbotapi.NewWebhook(bm.cfg.Telegram.WebhookURL)
	if err != nil {
		return fmt.Errorf("failed to create webhook config: %w", err)
	}
	_, err = bm.bot.Request(webhookConfig)
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	logger.Info("Webhook configured successfully", logrus.Fields{
		"webhook_url": bm.cfg.Telegram.WebhookURL,
	})

	return nil
}

func (bm *BotManager) RemoveWebhook() error {
	if bm.bot == nil {
		return fmt.Errorf("bot instance is nil")
	}

	_, err := bm.bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		return fmt.Errorf("failed to remove webhook: %w", err)
	}

	logger.Info("Webhook removed successfully")
	return nil
}