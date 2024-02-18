package irregular_verbs

import (
	utilService "Yulia-Lingo/internal/util_services"

	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const IrregularVerbsCountPerPage = 5

type KeyboardVerbValue struct {
	Request string
	Page    int
	Latter  string
}

func HandleIrregularVerbListCallback(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error {
	callbackData := callbackQuery.Data
	callbackChatId := callbackQuery.Message.Chat.ID

	keyboardVerbValue, err := KeyboardVerbValueFromJSON(callbackData)
	if err != nil {
		return fmt.Errorf("failed to map keyboardVerbValue: %v", err)
	}
	selectedLetter := keyboardVerbValue.Latter
	currentPageNumber := keyboardVerbValue.Page

	irregularVerbsPageAsText, err := GetIrregularVerbsPageAsText(currentPageNumber, selectedLetter)
	if err != nil {
		return fmt.Errorf("failed to get irregular irregularVerbs page as text: %v", err)
	}

	var responseText string
	if irregularVerbsPageAsText != "" {
		responseText = utilService.GetMessageDelimiter() + "\n" +
			fmt.Sprintf("*Список неправильных глаголов на букву '%s':*\n\n", selectedLetter) +
			irregularVerbsPageAsText
	} else {
		responseText = utilService.GetMessageDelimiter() + "\n" +
			fmt.Sprintf("*Список неправильных глаголов на букву '%s' пуст*", selectedLetter)
	}

	keyboard, err := CreateInlineKeyboard(keyboardVerbValue.Page, selectedLetter)
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

func GetIrregularVerbsPageAsText(currentPageNumber int, selectedLetter string) (string, error) {
	offset := currentPageNumber * IrregularVerbsCountPerPage
	irregularVerbsListPage, err := GetIrregularVerbsListPage(offset, IrregularVerbsCountPerPage, selectedLetter)
	if err != nil {
		return "", fmt.Errorf("failed to get irregularVerbs page from database: %v", err)
	}
	if len(irregularVerbsListPage) == 0 {
		return "", nil
	}
	var irregularVerbsPageAsText string
	for _, verb := range irregularVerbsListPage {
		irregularVerbsPageAsText += fmt.Sprintf("*%s*:\n*[*%s / %s / %s*]*\n\n", verb.Original, verb.Verb, verb.Past, verb.PastParticiple)
	}
	return irregularVerbsPageAsText, nil
}

func CreateInlineKeyboard(currentPage int, letter string) ([]tgbotapi.InlineKeyboardButton, error) {
	totalVerbs, err := GetTotalIrregularVerbsCount(letter)
	if err != nil {
		return nil, fmt.Errorf("failed to get total irregular verbs count: %v", err)
	}
	totalPages := totalVerbs / IrregularVerbsCountPerPage

	var keyboard []tgbotapi.InlineKeyboardButton
	if currentPage > 0 {
		jsonPrev, err := utilService.ConvertToJson(KeyboardVerbValue{
			Request: "IrregularVerbs",
			Page:    currentPage - 1,
			Latter:  letter,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create a json for the case (currentPage > 0): %v", err)
		}
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardButtonData("⬅️Назад", jsonPrev))
	}
	if currentPage < totalPages && totalVerbs > IrregularVerbsCountPerPage {
		jsonNext, err := utilService.ConvertToJson(KeyboardVerbValue{
			Request: "IrregularVerbs",
			Page:    currentPage + 1,
			Latter:  letter,
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

func KeyboardVerbValueFromJSON(jsonStr string) (KeyboardVerbValue, error) {
	var kv KeyboardVerbValue
	err := json.Unmarshal([]byte(jsonStr), &kv)
	if err != nil {
		return KeyboardVerbValue{}, err
	}
	return kv, nil
}
