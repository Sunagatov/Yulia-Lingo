package button

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func HandleSaveWordButtonClick(bot *tgbotapi.BotAPI, chatID int64) {
	responseText := "Слово сохранено для последующего изучения"
	msg := tgbotapi.NewMessage(chatID, responseText)
	_, errorMessage := bot.Send(msg)
	if errorMessage != nil {
		log.Printf("Error sending response message: %v", errorMessage)
	}
}
