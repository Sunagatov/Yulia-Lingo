package bot_manager

import (
	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/logger"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func ConfigureTelegramBotWebhook(bot *tgbotapi.BotAPI, cfg *config.Config) error {
	if cfg.Telegram.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	webhook, err := tgbotapi.NewWebhook(cfg.Telegram.WebhookURL)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	if _, err := bot.Request(webhook); err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	logger.Info("Webhook configured successfully", logrus.Fields{
		"webhook_url": cfg.Telegram.WebhookURL,
	})

	return nil
}

func RemoveWebhook(bot *tgbotapi.BotAPI) error {
	removeWebhook := tgbotapi.DeleteWebhookConfig{}
	if _, err := bot.Request(removeWebhook); err != nil {
		return fmt.Errorf("failed to remove webhook: %w", err)
	}

	logger.Info("Webhook removed successfully")
	return nil
}
