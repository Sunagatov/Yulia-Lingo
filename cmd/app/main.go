package main

import (
	database "Yulia-Lingo/internal/db"
	"Yulia-Lingo/internal/server"
	tg "Yulia-Lingo/internal/telegram"
	"Yulia-Lingo/internal/telegram/handler"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	database.CreateDatabaseConnection()
	defer database.CloseDatabaseConnection()

	err := database.InitDatabase()
	if err != nil {
		log.Fatalf("Error database init: %v", err)
	}

	tg.CreateNewTelegramBot()

	err = tg.SetupTelegramBotWebhook()
	if err != nil {
		log.Fatalf("Error creating telegram bot webhook: %v", err)
	}

	err = server.StartHTTPServer()
	if err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}

	err = handler.HandleBotUpdates()
	if err != nil {
		log.Fatalf("Error starting handle  bot updates: %v", err)
	}
}
