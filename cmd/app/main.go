package main

import (
	database "Yulia-Lingo/internal/db"
	"Yulia-Lingo/internal/telegram/handler"
	verb_database "Yulia-Lingo/internal/verb/db"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

func main() {
	database.CreateDatabaseConnection()
	defer database.CloseDatabaseConnection()

	err := verb_database.InitIrregularVerbsTable()
	if err != nil {
		log.Fatalf("Error database init: %v", err)
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatalf("no TELEGRAM_BOT_TOKEN provided in environment variables")
	}
	log.Println("botToken: " + botToken)

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	webhookURL := os.Getenv("TELEGRAM_WEBHOOK_URL")
	if webhookURL == "" {
		fmt.Errorf("no WEBHOOK_URL provided in environment variables")
	}

	log.Println("webhookURL: " + webhookURL)

	wh, _ := tgbotapi.NewWebhook(webhookURL)

	_, err = bot.Request(wh)
	if err != nil {
		log.Fatal(err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/" + bot.Token)
	go http.ListenAndServe("0.0.0.0:8083", nil)

	for update := range updates {
		if update.Message != nil {
			err = handler.HandleMessageFromUser(bot, update)
			if err != nil {
				log.Fatalf("Error handling message from a user: %v", err)
			}
		} else if update.CallbackQuery != nil {
			err = handler.HandleCallbackQuery(bot, update)
			if err != nil {
				log.Fatalf("Error handling callback from a user: %v", err)
			}
		}
	}
}
