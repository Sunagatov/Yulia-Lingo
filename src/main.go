package main

func main() {
	telegramBot := initTelegramBot()
	setupWebhook(telegramBot)
	startHTTPServer()
	listenForTelegramBotUpdates(telegramBot)
}
