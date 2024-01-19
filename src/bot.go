package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
)

func initBot() *tgbotapi.BotAPI {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("No TELEGRAM_BOT_TOKEN provided in environment variables")
	}

	bot, botInitError := tgbotapi.NewBotAPI(botToken)
	if botInitError != nil {
		log.Fatalf("Failed to authorize bot with token %s: %v", botToken, botInitError)
	}
	bot.Debug = true

	log.Printf("Authorized on Telegram account: %s", bot.Self.UserName)
	return bot
}
