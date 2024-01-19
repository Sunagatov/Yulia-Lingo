package main

func main() {
	bot := initBot()
	setupWebhook(bot)
	startHTTPServer()
	listenForUpdates(bot)
}
