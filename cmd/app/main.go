package main

import (
	database "Yulia-Lingo/internal/db"
	"Yulia-Lingo/internal/server"
	tg "Yulia-Lingo/internal/telegram"
	"Yulia-Lingo/internal/telegram/handler"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

func main() {
	db, err := database.CreateDatabaseConnection()
	if err != nil {
		log.Printf("Error opening database: %v", err)
		panic(fmt.Sprintf("Error opening database: %v", err))
	}

	err = database.InitDatabase(db)
	if err != nil {
		log.Printf("Error database init: %v", err)
		panic(fmt.Sprintf("Error database init: %v", err))
	}

	bot, err := tg.CreateNewTelegramBot()
	if err != nil {
		log.Printf("Error creating telegram bot: %v", err)
		panic(fmt.Sprintf("Error creating telegram bot: %v", err))
	}

	err = tg.SetupTelegramBotWebhook(bot)
	if err != nil {
		log.Printf("Error creating telegram bot webhook: %v", err)
		panic(fmt.Sprintf("Error creating telegram bot webhook: %v", err))
	}

	err = server.StartHTTPServer()
	if err != nil {
		log.Printf("Error starting HTTP server: %v", err)
		panic(fmt.Sprintf("Error starting HTTP server: %v", err))
	}

	handler.HandleBotUpdates(bot, db)
}
