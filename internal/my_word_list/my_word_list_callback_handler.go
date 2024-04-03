package my_word_list

import (
	utilService "Yulia-Lingo/internal/util_services"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type KeyboardVerbValue struct {
	Req          string
	Page         int
	PartOfSpeech string
}

func KeyboardVerbValueFromJSON(jsonStr string) (KeyboardVerbValue, error) {
	var kv KeyboardVerbValue
	err := json.Unmarshal([]byte(jsonStr), &kv)
	if err != nil {
		return KeyboardVerbValue{}, err
	}
	return kv, nil
}

const MyWordsListCountPerPage = 5

func HandleIrregularVerbListCallback(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error {
	callbackData := callbackQuery.Data
	callbackChatId := callbackQuery.Message.Chat.ID

	keyboardVerbValue, err := KeyboardVerbValueFromJSON(callbackData)
	if err != nil {
		return fmt.Errorf("failed to map keyboardVerbValue: %v", err)
	}

	partsOfSpeech := map[string]string{
		"N":      "Noun",
		"V":      "Verb",
		"Adj":    "Adjective",
		"Adv":    "Adverb",
		"Pro":    "Pronoun",
		"Prep":   "Preposition",
		"Conj":   "Conjunction",
		"Interj": "Interjection",
	}

	partOfSpeech := partsOfSpeech[keyboardVerbValue.PartOfSpeech]
	currentPageNumber := keyboardVerbValue.Page

	myWordListPageAsText, err := GetMyWordListPageAsText(currentPageNumber, partOfSpeech)
	if err != nil {
		return fmt.Errorf("failed to get irregular irregularVerbs page as text: %v", err)
	}

	var responseText string
	if myWordListPageAsText != "" {
		responseText = utilService.GetMessageDelimiter() + "\n" +
			fmt.Sprintf("*Список глаголов:*\n\n") +
			myWordListPageAsText
	} else {
		responseText = utilService.GetMessageDelimiter() + "\n" +
			fmt.Sprintf("*Список глаголов пуст*")
	}

	keyboard, err := CreateInlineKeyboard(keyboardVerbValue.Page, partOfSpeech)
	if err != nil {
		return fmt.Errorf("failed to inline keyboard: %v", err)
	}

	messageToUser := tgbotapi.NewMessage(callbackChatId, responseText)
	messageToUser.ParseMode = "Markdown"
	if keyboard != nil {
		messageToUser.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard)
	}

	_, err = bot.Send(&messageToUser)
	if err != nil {
		return fmt.Errorf("failed to send response message: %v", err)
	}
	return nil
}

func GetMyWordListPageAsText(currentPageNumber int, partOfSpeech string) (string, error) {
	offset := currentPageNumber * MyWordsListCountPerPage
	myWordListPage, err := GetEnglishWordsWithRussianTranslations(offset, MyWordsListCountPerPage, partOfSpeech)
	if err != nil {
		return "", fmt.Errorf("failed to get irregularVerbs page from database: %v", err)
	}
	if len(myWordListPage) == 0 {
		return "", nil
	}
	var irregularVerbsPageAsText string
	for _, word := range myWordListPage {
		irregularVerbsPageAsText += fmt.Sprintf("*%s*:\n*[*%s*]*\n\n", word.EnglishWord, word.Translations)
	}
	return irregularVerbsPageAsText, nil
}

func CreateInlineKeyboard(currentPage int, partOfSpeech string) ([]tgbotapi.InlineKeyboardButton, error) {
	totalVerbs, err := GetTotalMyWordsListCount(partOfSpeech)
	if err != nil {
		return nil, fmt.Errorf("failed to get total my word list count: %v", err)
	}
	totalPages := totalVerbs / MyWordsListCountPerPage

	var keyboard []tgbotapi.InlineKeyboardButton
	if currentPage > 0 {
		jsonPrev, err := utilService.ConvertToJson(KeyboardVerbValue{
			Req:          "MyWordList",
			Page:         currentPage - 1,
			PartOfSpeech: partOfSpeech,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create a json for the case (currentPage > 0): %v", err)
		}
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardButtonData("⬅️Назад", jsonPrev))
	}
	if currentPage < totalPages && totalVerbs > MyWordsListCountPerPage {
		jsonNext, err := utilService.ConvertToJson(KeyboardVerbValue{
			Req:          "IrregularVerbs",
			Page:         currentPage + 1,
			PartOfSpeech: partOfSpeech,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create a json for the case (currentPage < totalPages): %v", err)
		}
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardButtonData("Вперед ➡️", jsonNext))
	}

	if len(keyboard) == 0 {
		return nil, nil
	}

	return keyboard, nil
}
