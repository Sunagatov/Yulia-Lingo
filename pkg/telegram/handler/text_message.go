package handler

import (
	"Yulia-Lingo/internal/message_sender"
	"Yulia-Lingo/internal/message_sender/model"
	"Yulia-Lingo/internal/word/api"
	"github.com/spf13/viper"
	"log"
	"regexp"
)

func (handle *MessageHandler) TextMessage() {
	messageFromUser := handle.Update.Message.Text

	message := model.Message{
		TelegramBot: handle.TelegramBot,
		ChatId:      handle.Update.Message.Chat.ID,
		Text:        "",
		Keyboard:    nil,
	}

	if !isValidWord(messageFromUser) {
		message.Text = "Please send a single, valid word in English."
		message_sender.SendMessageToUser(message)
		return
	}

	responseMessage, err := api.RequestWordsAPI(messageFromUser)
	message.Text = responseMessage

	if err != nil {
		log.Printf("Error fetching from API: %v", err)
		message.Text = "Sorry, there was an error processing your request."
		message_sender.SendMessageToUser(message)
		return
	}

	addKeyboardButton(message)
	message_sender.SendMessageToUser(message)
}

func addKeyboardButton(message model.Message) {
	message.Keyboard = []model.KeyboardButton{
		{Text: viper.GetString("buttons.save-word.key"), CallbackData: viper.GetString("buttons.save-word.value")},
	}
}

func isValidWord(word string) bool {
	const pattern = `^[A-Za-z]+$`
	matched, _ := regexp.MatchString(pattern, word)
	return matched
}
