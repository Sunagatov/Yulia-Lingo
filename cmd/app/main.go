package main

import (
	"Yulia-Lingo/internal/config/server"
	tg "Yulia-Lingo/pkg/telegram"
	tg_config "Yulia-Lingo/pkg/telegram/config"
)

func main() {
	telegramBot := tg_config.CreateNewTelegramBot()
	tg_config.SetupTelegramBotWebhook(telegramBot)
	server.StartHTTPServer()
	tg.LaunchTelegramBot(telegramBot)
}
