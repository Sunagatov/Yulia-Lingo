package button

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func HandleStartButtonClick(bot *tgbotapi.BotAPI, chatID int64) {
	userName := bot.Self.FirstName
	text := fmt.Sprintf(GreetingMessageToUser, userName)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(IrregularVerbListButtonName)),
	)
	_, errorMessage := bot.Send(msg)
	if errorMessage != nil {
		log.Printf("Error sending response message: %v", errorMessage)
	}
}
