package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
)

func initTelegramBot() *tgbotapi.BotAPI {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("No TELEGRAM_BOT_TOKEN provided in environment variables")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Failed to authorize Telegram bot with token %s: %v", botToken, err)
	}
	bot.Debug = true

	log.Printf("Authorized on Telegram account: %s", bot.Self.UserName)
	return bot
}
