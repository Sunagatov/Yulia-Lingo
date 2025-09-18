package my_word_list

import (
	utilService "Yulia-Lingo/internal/util_services"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleMyWordListButtonClick(bot *tgbotapi.BotAPI, chatID int64) error {
	inlineKeyboardMarkup, err := CreatePartsOfSpeechKeyboardMarkup()
	if err != nil {
		return fmt.Errorf("failed to create inlineKeyboardMarkup for MyWordList: %v", err)
	}

	messageText := "*Слова какой части речи вы хотите посмотреть?*\n\n"
	messageToUser := tgbotapi.NewMessage(chatID, messageText)
	messageToUser.ParseMode = "Markdown"
	messageToUser.ReplyMarkup = inlineKeyboardMarkup

	_, err = bot.Send(&messageToUser)
	if err != nil {
		return fmt.Errorf("failed to send the message for 'MyWordList' button to a user: %v", err)
	}
	return nil
}

func CreatePartsOfSpeechKeyboardMarkup() (*tgbotapi.InlineKeyboardMarkup, error) {
	var rows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	partsOfSpeech := getPartsOfSpeechRussianMapping()

	for abbreviation, russianName := range partsOfSpeech {
		requestData := KeyboardRequestData{
			Req:          "MyWordList",
			Page:         0,
			PartOfSpeech: abbreviation,
		}
		jsonAsString, err := utilService.ConvertToJson(requestData)
		if err != nil {
			return nil, fmt.Errorf("failed to create JSON for partOfSpeech '%s': %w", abbreviation, err)
		}
		btn := tgbotapi.NewInlineKeyboardButtonData(russianName, jsonAsString)
		currentRow = append(currentRow, btn)

		if len(currentRow) == 2 {
			rows = append(rows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}, nil
}

type KeyboardRequestData struct {
	Req          string
	Page         int
	PartOfSpeech string
}

func getPartsOfSpeechRussianMapping() map[string]string {
	return map[string]string{
		"N":      "Существительное",
		"V":      "Глагол",
		"Adj":    "Прилагательное",
		"Adv":    "Наречие",
		"Pro":    "Местоимение",
		"Prep":   "Предлог",
		"Conj":   "Союз",
		"Interj": "Междометие",
	}
}
