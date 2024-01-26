package setup

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SetupTelegramBotWebhook(telegramBot *tgbotapi.BotAPI) {
	webhookURL := os.Getenv("WEBHOOK_URL") + "/" + telegramBot.Token
	if webhookURL == "" {
		log.Fatal("No WEBHOOK_URL provided in environment variables")
		return
	}

	webhookConfig, webhookConfigError := tgbotapi.NewWebhook(webhookURL)
	if webhookConfigError != nil {
		log.Fatalf("Failed to create webhook: %v", webhookConfigError)
		return
	}

	_, webhookSetError := telegramBot.Request(webhookConfig)
	if webhookSetError != nil {
		log.Fatalf("Failed to set webhook: %v", webhookSetError)
		return
	}

	info, webhookInfoError := telegramBot.GetWebhookInfo()
	if webhookInfoError != nil {
		log.Fatalf("Failed to get webhook info: %v", webhookInfoError)
		return
	}
	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	} else {
		log.Printf("Webhook successfully set to %s", webhookURL)
	}
}
