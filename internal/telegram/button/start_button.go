package button

import (
	"Yulia-Lingo/internal/telegram/message"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
)

func HandleStartButtonClick(chatID int64) {
	userName := os.Getenv("TG_BOT_NAME")

	text := fmt.Sprintf(GreetingMessageToUser, userName)
	msg := tgbotapi.NewMessage(chatID, text)
	addKeyboardButton(&msg)
	errorMessage := message.Send(&msg)

	if errorMessage != nil {
		log.Printf("Error sending response message: %v", errorMessage)
	}
}

func addKeyboardButton(msg *tgbotapi.MessageConfig) {
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(IrregularVerbListButtonName)),
	)
}
