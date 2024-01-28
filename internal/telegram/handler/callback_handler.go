package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func HandleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	callbackChatID := callbackQuery.Message.Chat.ID
	callbackMessageID := callbackQuery.Message.MessageID
	callbackMessageText := callbackQuery.Message.Text

	switch callbackQuery.Data {
	case "save_word_option":
		msg := tgbotapi.NewEditMessageText(callbackChatID, callbackMessageID, callbackMessageText)
		bot.Send(msg)

		responseText := fmt.Sprintf("Слово '%s' сохранено для последующего изучения", callbackMessageText)
		messageToUser := tgbotapi.NewMessage(callbackChatID, responseText)
		_, errorMessage := bot.Send(messageToUser)
		if errorMessage != nil {
			log.Printf("Error sending response message: %v", errorMessage)
		}
	case "📘 Мой список слов":
		responseText := "callbackQuery Список слов пока пуст"
		callbackMessage := tgbotapi.NewEditMessageText(callbackChatID, callbackMessageID, responseText)
		bot.Send(callbackMessage)
	default:
		responseText := "Эта функция пока что в работе и не поддерживается"
		callbackMessage := tgbotapi.NewEditMessageText(callbackChatID, callbackMessageID, responseText)
		bot.Send(callbackMessage)
	}
}
