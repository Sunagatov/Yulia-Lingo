package main

import (
	"Yulia-Lingo/internal/server"
	tg "Yulia-Lingo/internal/telegram"
)

func main() {
	telegramBot := tg.CreateNewTelegramBot()
	tg.SetupTelegramBotWebhook(telegramBot)
	server.StartHTTPServer()
	tg.LaunchTelegramBot(telegramBot)
}
