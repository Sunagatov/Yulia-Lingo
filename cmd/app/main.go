package main

import (
	"Yulia-Lingo/internal/server"
	tg "Yulia-Lingo/internal/telegram"
	"Yulia-Lingo/internal/telegram/handler"
)

func main() {
	bot := tg.CreateNewTelegramBot()
	tg.SetupTelegramBotWebhook(bot)
	server.StartHTTPServer()
	handler.HandleBotUpdates(bot)
}
