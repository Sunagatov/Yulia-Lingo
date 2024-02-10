package handler

import (
	"Yulia-Lingo/internal/telegram/button"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"math"
	"strings"
)

func HandleCallbackQuery(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update) {
	callbackQuery := botUpdate.CallbackQuery

	callbackChatID := callbackQuery.Message.Chat.ID
	callbackMessageID := callbackQuery.Message.MessageID
	callbackMessageText := callbackQuery.Message.Text
	callbackData := callbackQuery.Data

	switch {
	case strings.HasPrefix(callbackQuery.Data, "select_letter_"):
		selectedLetter := strings.TrimPrefix(callbackData, "select_letter_")
		responseText := fmt.Sprintf("Список неправильных глаголов на букву '%s':\n\n", selectedLetter)

		pageNumber := 1
		button.UpdateCurrentPage(callbackChatID, pageNumber)
		currentPage := button.GetCurrentPage(callbackChatID)
		totalVerbs, err := button.GetTotalIrregularVerbsCount()
		if err != nil {
			log.Printf("Error getting total irregular verbs count: %v", err)
			return
		}
		totalPages := int(math.Ceil(float64(totalVerbs) / button.IrregularVerbsPerPage))

		offset := (currentPage - 1) * button.IrregularVerbsPerPage
		verbs, err := button.GetIrregularVerbs(offset, button.IrregularVerbsPerPage, selectedLetter)
		if err != nil {
			log.Printf("Error getting irregular verbs: %v", err)
			return
		}

		var messageText string
		for _, verb := range verbs {
			messageText += fmt.Sprintf("%s - [%s - %s - %s]\n", verb.Original, verb.Verb, verb.Past, verb.PastP)
		}

		responseText = responseText + messageText

		log.Printf("pageNumber: %s,\n\n currentPage: %s,\n\n totalVerbs: %s,\n\n totalPages: %s,\n\n offset: %s,\n\n verbs: %s,\n\n responseText: %s\n\n",
			pageNumber, currentPage, totalVerbs, totalPages, offset, verbs, responseText)

		messageToUser := tgbotapi.NewMessage(callbackChatID, responseText)
		messageToUser.ReplyMarkup = button.CreateInlineKeyboard(currentPage, totalPages, selectedLetter)

		_, errorMessage := bot.Send(&messageToUser)
		if errorMessage != nil {
			log.Printf("Error sending response message: %v", errorMessage)
		}

	case strings.HasPrefix(callbackQuery.Data, "irregular_verbs_page_"):
		pageNumber, letter := button.ExtractPageNumber(callbackData)

		button.UpdateCurrentPage(callbackChatID, pageNumber)

		msg := tgbotapi.NewEditMessageText(callbackChatID, callbackMessageID, callbackMessageText)
		_, err := bot.Send(&msg)
		if err != nil {
			log.Printf("Error with edit message, err: %v", err)
		}

		currentPage := button.GetCurrentPage(callbackChatID)

		totalVerbs, err := button.GetTotalIrregularVerbsCount()
		if err != nil {
			log.Printf("Error getting total irregular verbs count: %v", err)
			return
		}
		totalPages := int(math.Ceil(float64(totalVerbs) / button.IrregularVerbsPerPage))

		offset := (currentPage - 1) * button.IrregularVerbsPerPage
		verbs, err := button.GetIrregularVerbs(offset, button.IrregularVerbsPerPage, letter)
		if err != nil {
			log.Printf("Error getting irregular verbs: %v", err)
			return
		}

		var messageText string
		for _, verb := range verbs {
			messageText += fmt.Sprintf("%s - [%s - %s - %s]\n", verb.Original, verb.Verb, verb.Past, verb.PastP)
		}

		messageToUser := tgbotapi.NewMessage(callbackChatID, messageText)
		messageToUser.ReplyMarkup = button.CreateInlineKeyboard(currentPage, totalPages, letter)

		_, errorMessage := bot.Send(&messageToUser)
		if errorMessage != nil {
			log.Printf("Error sending response message: %v", errorMessage)
		}

	default:
		responseText := "Эта функция пока что в работе и не поддерживается"
		callbackMessage := tgbotapi.NewEditMessageText(callbackChatID, callbackMessageID, responseText)
		_, err := bot.Send(callbackMessage)
		if err != nil {
			log.Printf("Error with edit message, err: %v", err)
		}
	}
}
