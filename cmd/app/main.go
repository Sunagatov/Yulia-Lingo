package main

import (
	"Yulia-Lingo/internal/server"
	envFileLoader "Yulia-Lingo/pkg/common"
	tg "Yulia-Lingo/pkg/telegram"
)

func main() {
	envFileLoader.Load()
	telegramBot := tg.CreateNewTelegramBot()
	tg.SetupTelegramBotWebhook(telegramBot)
	server.StartHTTPServer()
	tg.LaunchTelegramBot(telegramBot)
}
