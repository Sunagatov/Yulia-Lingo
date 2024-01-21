package main

import (
	"Yulia-Lingo/internal/server"
	tg "Yulia-Lingo/pkg/telegram"
)

func main() {
	telegramBot := tg.CreateNewTelegramBot()
	tg.SetupTelegramBotWebhook(telegramBot)
	server.StartHTTPServer()
	tg.LaunchTelegramBot(telegramBot)
}
