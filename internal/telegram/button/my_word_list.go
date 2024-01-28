package button

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func HandleMyWordListButtonClick(bot *tgbotapi.BotAPI, chatID int64) {
	text := "Выберите часть речи"
	msg := tgbotapi.NewMessage(chatID, text)

	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(VerbsButtonName)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("🟣Существительные")),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("🟠Прилагательные")),
	)
	_, errorMessage := bot.Send(msg)
	if errorMessage != nil {
		log.Printf("Error sending response message: %v", errorMessage)
	}
}
