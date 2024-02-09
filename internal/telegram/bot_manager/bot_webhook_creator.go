package bot_manager

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
)

func CreateTelegramWebhook(err error, bot *tgbotapi.BotAPI) error {
	webhookURL := os.Getenv("TELEGRAM_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("No WEBHOOK_URL provided in environment variables")
	} else {
		log.Println("WebhookURL: " + webhookURL)
	}

	wh, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		return fmt.Errorf("Error creating webhook: %v", err)
	}

	_, err = bot.Request(wh)
	if err != nil {
		return fmt.Errorf("Error setting webhook: %v", err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		return fmt.Errorf("Error setting webhook: %v", err)
	}

	if info.LastErrorDate != 0 {
		return fmt.Errorf("Telegram callback failed: %s", info.LastErrorMessage)
	}
	return nil
}
