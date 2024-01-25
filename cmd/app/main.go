package main

import (
	"Yulia-Lingo/internal/db"
	"Yulia-Lingo/internal/server"
	"Yulia-Lingo/pkg/common"
	tg "Yulia-Lingo/pkg/telegram"
)

func main() {
	common.Init()

	connectDB := db.CreateDatabaseConnection()
	db.UpMigrations(connectDB)

	telegramBot := tg.CreateNewTelegramBot()
	tg.SetupTelegramBotWebhook(telegramBot)
	server.StartHTTPServer()
	tg.LaunchTelegramBot(telegramBot)
}
