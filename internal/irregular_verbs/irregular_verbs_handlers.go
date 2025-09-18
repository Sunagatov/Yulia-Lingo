package irregular_verbs

import (
	"encoding/json"
	"fmt"

	"Yulia-Lingo/internal/util"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const IrregularVerbsCountPerPage = 5

func HandleIrregularVerbListCallback(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error {
	if callbackQuery == nil || callbackQuery.Message == nil {
		return fmt.Errorf("invalid callback query")
	}

	callbackData := callbackQuery.Data
	chatID := callbackQuery.Message.Chat.ID

	var keyboardValue KeyboardVerbValue
	if err := json.Unmarshal([]byte(callbackData), &keyboardValue); err != nil {
		return fmt.Errorf("failed to unmarshal callback data: %w", err)
	}

	selectedLetter := keyboardValue.Letter
	currentPage := keyboardValue.Page

	pageText, err := getIrregularVerbsPageText(currentPage, selectedLetter)
	if err != nil {
		return fmt.Errorf("failed to get page text: %w", err)
	}

	responseText := util.GetMessageDelimiter() + "\n"
	if pageText != "" {
		responseText += fmt.Sprintf("*Список неправильных глаголов на букву '%s':*\n\n%s", selectedLetter, pageText)
	} else {
		responseText += fmt.Sprintf("*Список неправильных глаголов на букву '%s' пуст*", selectedLetter)
	}

	keyboard, err := createInlineKeyboard(currentPage, selectedLetter)
	if err != nil {
		return fmt.Errorf("failed to create keyboard: %w", err)
	}

	msg := tgbotapi.NewMessage(chatID, responseText)
	msg.ParseMode = "Markdown"
	if keyboard != nil {
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(*keyboard)
	}

	_, err = bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func getIrregularVerbsPageText(currentPage int, selectedLetter string) (string, error) {
	offset := currentPage * IrregularVerbsCountPerPage
	
	// This would need to be injected properly in a real refactor
	// For now, using the old function to maintain compatibility
	entities, err := GetIrregularVerbsListPage(offset, IrregularVerbsCountPerPage, selectedLetter)
	if err != nil {
		return "", fmt.Errorf("failed to get page from database: %w", err)
	}

	if len(entities) == 0 {
		return "", nil
	}

	var result string
	for _, verb := range entities {
		result += fmt.Sprintf("*%s*:\n*[*%s / %s / %s*]*\n\n", verb.Original, verb.Verb, verb.Past, verb.PastParticiple)
	}

	return result, nil
}

func createInlineKeyboard(currentPage int, letter string) (*[]tgbotapi.InlineKeyboardButton, error) {
	totalVerbs, err := GetTotalIrregularVerbsCount(letter)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	totalPages := (totalVerbs + IrregularVerbsCountPerPage - 1) / IrregularVerbsCountPerPage

	var buttons []tgbotapi.InlineKeyboardButton

	if currentPage > 0 {
		prevData := KeyboardVerbValue{
			Request: "IrregularVerbs",
			Page:    currentPage - 1,
			Letter:  letter,
		}
		jsonPrev, err := util.ConvertToJSON(prevData)
		if err != nil {
			return nil, fmt.Errorf("failed to create JSON for previous button: %w", err)
		}
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", jsonPrev))
	}

	if currentPage < totalPages-1 && totalVerbs > IrregularVerbsCountPerPage {
		nextData := KeyboardVerbValue{
			Request: "IrregularVerbs",
			Page:    currentPage + 1,
			Letter:  letter,
		}
		jsonNext, err := util.ConvertToJSON(nextData)
		if err != nil {
			return nil, fmt.Errorf("failed to create JSON for next button: %w", err)
		}
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Вперед ➡️", jsonNext))
	}

	if len(buttons) == 0 {
		return nil, nil
	}

	return &buttons, nil
}