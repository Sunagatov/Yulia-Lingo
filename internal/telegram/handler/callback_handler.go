package handler

import (
	"Yulia-Lingo/internal/verb/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

func HandleCallbackQuery(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update) {
	callbackQuery := botUpdate.CallbackQuery

	callbackChatID := callbackQuery.Message.Chat.ID
	callbackMessageID := callbackQuery.Message.MessageID
	callbackData := callbackQuery.Data

	switch {
	case strings.Contains(callbackData, "GetListByLatter"):
		service.GetVerbsListByLatter(callbackQuery, bot)
	default:
		responseText := "Эта функция пока что в работе и не поддерживается"
		callbackMessage := tgbotapi.NewEditMessageText(callbackChatID, callbackMessageID, responseText)
		_, err := bot.Send(callbackMessage)
		if err != nil {
			log.Printf("Error with edit message, err: %v", err)
		}
	}
}
