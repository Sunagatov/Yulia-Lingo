package main

import (
	"Yulia-Lingo/internal/db"
	"Yulia-Lingo/internal/server"
	"Yulia-Lingo/pkg/common"
	tgHandler "Yulia-Lingo/pkg/telegram/handler"
	tg "Yulia-Lingo/pkg/telegram/set_up"
)

func main() {
	common.Init()

	connectDB := db.CreateDatabaseConnection()
	db.UpMigrations(connectDB)
	defer db.DB.Close()

	telegramBot := tg.CreateNewTelegramBot()
	tg.SetupTelegramBotWebhook(telegramBot)
	server.StartHTTPServer()
	tgHandler.LaunchTelegramBot(telegramBot)
}
