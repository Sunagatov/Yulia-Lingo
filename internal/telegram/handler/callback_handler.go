package handler

import (
	"Yulia-Lingo/internal/telegram/button"
	"Yulia-Lingo/internal/telegram/message"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

func HandleCallbackQuery(botUpdate tgbotapi.Update) {
	callbackQuery := botUpdate.CallbackQuery

	callbackChatID := callbackQuery.Message.Chat.ID
	callbackMessageID := callbackQuery.Message.MessageID
	callbackMessageText := callbackQuery.Message.Text
	callbackData := callbackQuery.Data

	switch {
	case strings.HasPrefix(callbackQuery.Data, "irregular_verbs_page_"):
		pageNumber := button.ExtractPageNumber(callbackData)

		// Update the current page in user's context
		button.UpdateCurrentPage(callbackChatID, pageNumber)

		msg := tgbotapi.NewEditMessageText(callbackChatID, callbackMessageID, callbackMessageText)
		editMessage(&msg)

		// Handle the Irregular Verbs button click
		button.HandleIrregularVerbsListButtonClick(callbackChatID)

	default:
		responseText := "Эта функция пока что в работе и не поддерживается"
		callbackMessage := tgbotapi.NewEditMessageText(callbackChatID, callbackMessageID, responseText)
		editMessage(&callbackMessage)
	}
}

func editMessage(msg *tgbotapi.EditMessageTextConfig) {
	err := message.Edit(msg)
	if err != nil {
		log.Printf("Error with edit message, err: %v", err)
	}
}
