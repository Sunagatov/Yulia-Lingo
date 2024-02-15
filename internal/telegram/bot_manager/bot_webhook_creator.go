package bot_manager

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
)

func ConfigureTelegramBotWebhook(bot *tgbotapi.BotAPI) error {
	webhookURL := os.Getenv("TELEGRAM_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("failed to read TELEGRAM_WEBHOOK_URL from the environment variables")
	}

	webhook, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %v", err)
	}

	_, err = bot.Request(webhook)
	if err != nil {
		return fmt.Errorf("failed to set webhook: %v", err)
	}

	_, err = bot.GetWebhookInfo()
	if err != nil {
		return fmt.Errorf("failed to get webhook info: %v", err)
	}

	return nil
}
