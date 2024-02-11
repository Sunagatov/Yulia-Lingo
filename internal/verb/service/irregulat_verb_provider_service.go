package service

import (
	"Yulia-Lingo/internal/common/service"
	"Yulia-Lingo/internal/verb/model"
	"Yulia-Lingo/internal/verb/repository"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

const ListLimit = 10

func GetVerbsListByLatter(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {
	callbackData := callbackQuery.Data

	keyboardVerbValue, err := model.KeyboardVerbValueFromJSON(callbackData)
	if err != nil {
		log.Printf("Can't map keyboardVerbValue, err: %v", err)
		return
	}

	letter := keyboardVerbValue.Latter
	responseText := fmt.Sprintf("Список неправильных глаголов на букву '%s':\n\n", letter)

	verbs, err := repository.GetVerbsListFromLatter(letter, keyboardVerbValue.Page, ListLimit)
	if err != nil {
		log.Printf("Can't get verb's list, err: %v", err)
		return
	}

	var messageText string
	for _, verb := range verbs {
		messageText += fmt.Sprintf("%s - [%s - %s - %s]\n", verb.Original, verb.Verb, verb.Past, verb.PastP)
	}

	responseText = responseText + messageText

	totalPage := getTotalPage(letter)
	messageToUser := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, responseText)
	createInlineKeyboard(&messageToUser, keyboardVerbValue.Page, totalPage, letter)

	_, errorMessage := bot.Send(&messageToUser)
	if errorMessage != nil {
		log.Printf("Error sending response message: %v", errorMessage)
	}
}

func createInlineKeyboard(messageToUser *tgbotapi.MessageConfig, currentPage, totalPages int64, letter string) {
	var keyboard []tgbotapi.InlineKeyboardButton
	if currentPage > 0 {
		jsonPrev := service.ToJSON(model.KeyboardVerbValue{
			Request: "GetListByLatter",
			Page:    currentPage - 1,
			Latter:  letter,
		})
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardButtonData("Prev page", jsonPrev))
	}
	if currentPage < totalPages {
		jsonNext := service.ToJSON(model.KeyboardVerbValue{
			Request: "GetListByLatter",
			Page:    currentPage + 1,
			Latter:  letter,
		})
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardButtonData("Next page", jsonNext))
	}

	if len(keyboard) == 0 {
		return
	}

	messageToUser.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard)
}

func getTotalPage(letter string) int64 {
	totalVerbs, err := repository.GetTotalIrregularVerbsCount(letter)
	if err != nil {
		log.Printf("Error getting total irregular verbs count: %v", err)
		return 0
	}
	return int64(totalVerbs / ListLimit)
}
