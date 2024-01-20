package main

func main() {
	telegramBot := createNewTelegramBot()
	setupTelegramBotWebhook(telegramBot)
	startHTTPServer()
	launchTelegramBot(telegramBot)
}
