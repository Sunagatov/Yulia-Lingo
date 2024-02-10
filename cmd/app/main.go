package main

import (
	dbManager "Yulia-Lingo/internal/db"
	botManager "Yulia-Lingo/internal/telegram/bot_manager"
	"Yulia-Lingo/internal/telegram/handler"
	"Yulia-Lingo/internal/verb/db"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

func main() {
	err := dbManager.CreateDatabaseConnection()
	if err != nil {
		log.Fatalf("Error creating database connection: %v", err)
	}
	defer dbManager.CloseDatabaseConnection()

	err = db.InitIrregularVerbsTable()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)

	} else {
		log.Println("Irregular verbs data inserted successfully.")
		log.Println("Database initialization completed successfully.")
	}

	bot, err := botManager.CreateTelegramBot()
	if err != nil {
		log.Fatalf("Error creating a new BotAPI instance: %v", err)
	} else {
		log.Println("A new BotAPI instance was created successfully.")
		log.Printf("Authorized on account %s", bot.Self.UserName)
		bot.Debug = true
	}

	err = botManager.CreateTelegramWebhook(bot)
	if err != nil {
		log.Fatalf("Error creating a new Telegram Bot Webhook: %v", err)
	} else {
		log.Println("Telegram Bot Webhook was created successfully.")
	}

	tgBotUpdates := bot.ListenForWebhook("/" + bot.Token)
	go startHTTPServer()

	for tgBotUpdate := range tgBotUpdates {
		if tgBotUpdate.Message != nil {
			handler.HandleMessageFromUser(bot, tgBotUpdate)
		} else if tgBotUpdate.CallbackQuery != nil {
			handler.HandleCallbackQuery(bot, tgBotUpdate)
		}
	}
}
func startHTTPServer() {
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		log.Fatalf("No APP_PORT provided in environment variables")
	}
	log.Printf("Starting HTTP server on port %s", appPort)
	if err := http.ListenAndServe("0.0.0.0:8083", nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
