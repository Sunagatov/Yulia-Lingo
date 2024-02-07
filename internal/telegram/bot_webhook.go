package telegram

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SetupTelegramBotWebhook() error {
	telegramBot, err := GetTelegramBot()
	if err != nil {
		return fmt.Errorf("app wosn't connect to telegram bot, err: %v", err)
	}

	webhookURL := os.Getenv("TELEGRAM_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("no WEBHOOK_URL provided in environment variables")
	}

	webhookConfig, webhookConfigError := tgbotapi.NewWebhook(webhookURL)
	if webhookConfigError != nil {
		return fmt.Errorf("failed to create webhook: %v", webhookConfigError)
	}

	_, webhookSetError := telegramBot.Request(webhookConfig)
	if webhookSetError != nil {
		return fmt.Errorf("failed to set webhook: %v", webhookSetError)
	}

	info, webhookInfoError := telegramBot.GetWebhookInfo()
	if webhookInfoError != nil {
		return fmt.Errorf("failed to get webhook info: %v", webhookInfoError)
	}

	if info.LastErrorDate != 0 {
		return fmt.Errorf("telegram callback failed: %s", info.LastErrorMessage)
	}

	log.Printf("Webhook successfully set to %s", webhookURL)
	return nil
}
