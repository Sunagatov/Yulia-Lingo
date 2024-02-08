package main

import (
	"Yulia-Lingo/internal/connections"
	database "Yulia-Lingo/internal/db"
	"Yulia-Lingo/internal/server"
	"Yulia-Lingo/internal/telegram/handler"
	"Yulia-Lingo/internal/telegram/setup"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	connections.Init()

	database.CreateDatabaseConnection()
	defer database.CloseDatabaseConnection()

	err := database.InitDatabase()
	if err != nil {
		log.Fatalf("Error database init: %v", err)
	}

	setup.CreateNewTelegramBot()
	err = setup.SetupTelegramBotWebhook()
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
