package handler

import (
	"Yulia-Lingo/internal/telegram/button"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

func HandleCallbackQuery(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update) {
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
		bot.Send(msg)

		// Handle the Irregular Verbs button click
		button.HandleIrregularVerbsListButtonClick(bot, callbackChatID)

	default:
		responseText := "Эта функция пока что в работе и не поддерживается"
		callbackMessage := tgbotapi.NewEditMessageText(callbackChatID, callbackMessageID, responseText)
		bot.Send(callbackMessage)
	}
}
