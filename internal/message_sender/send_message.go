package message_sender

import (
	"Yulia-Lingo/internal/message_sender/model"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const keyboardLen = 3

func SendMessageToUser(message model.Message) {
	bot := message.TelegramBot
	chatID := message.ChatId
	text := message.Text
	keyboard := prepareKeyboardButton(message.Keyboard)

	messageToUser := tgbotapi.NewMessage(chatID, text)
	messageToUser.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	_, err := bot.Send(messageToUser)
	if err != nil {
		log.Printf("Error sending message to a user: %v", err)
	}
}

func prepareKeyboardButton(inputKeyboardList []model.KeyboardButton) [][]tgbotapi.InlineKeyboardButton {
	var keyboard [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	for i, button := range inputKeyboardList {
		btn := tgbotapi.NewInlineKeyboardButtonData(button.Text, button.CallbackData)
		row = append(row, btn)
		if (i+1)%keyboardLen == 0 || i == len(inputKeyboardList)-1 {
			keyboard = append(keyboard, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}

	return keyboard
}
