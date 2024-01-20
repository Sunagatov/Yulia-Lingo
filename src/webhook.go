package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
)

func setupWebhook(telegramBot *tgbotapi.BotAPI) {
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatal("No WEBHOOK_URL provided in environment variables")
		return
	}
	fullWebhookURL := webhookURL + "/" + telegramBot.Token

	webhookConfig, webhookConfigError := tgbotapi.NewWebhook(fullWebhookURL)
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
		log.Printf("Webhook successfully set to %s", fullWebhookURL)
	}
}
