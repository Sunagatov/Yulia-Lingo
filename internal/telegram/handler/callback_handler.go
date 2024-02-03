package handler

import (
	"Yulia-Lingo/internal/telegram/button"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
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

	case callbackQuery.Data == "save_word_option":
		msg := tgbotapi.NewEditMessageText(callbackChatID, callbackMessageID, callbackMessageText)
		bot.Send(msg)

		responseText := fmt.Sprintf("Слово '%s' сохранено для последующего изучения", callbackMessageText)
		messageToUser := tgbotapi.NewMessage(callbackChatID, responseText)
		_, errorMessage := bot.Send(messageToUser)
		if errorMessage != nil {
			log.Printf("Error sending response message: %v", errorMessage)
		}
	case callbackQuery.Data == "📘 Мой список слов":
		responseText := "callbackQuery Список слов пока пуст"
		callbackMessage := tgbotapi.NewEditMessageText(callbackChatID, callbackMessageID, responseText)
		bot.Send(callbackMessage)
	default:
		responseText := "Эта функция пока что в работе и не поддерживается"
		callbackMessage := tgbotapi.NewEditMessageText(callbackChatID, callbackMessageID, responseText)
		bot.Send(callbackMessage)
	}
}

func handleIrregularVerbsPagination(bot *tgbotapi.BotAPI, chatID int64, messageID int, callbackData string) {

}
