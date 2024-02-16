package callback_handler

import (
	irregularVerbsManager "Yulia-Lingo/internal/database/irregular_verbs"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleIrregularVerbListCallback(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error {
	callbackData := callbackQuery.Data

	keyboardVerbValue, err := irregularVerbsManager.KeyboardVerbValueFromJSON(callbackData)
	if err != nil {
		return fmt.Errorf("failed to map keyboardVerbValue: %v", err)
	}

	selectedLetter := keyboardVerbValue.Latter
	currentPageNumber := keyboardVerbValue.Page

	irregularVerbsPageTitle := fmt.Sprintf("Список неправильных глаголов на букву '%s':\n\n", selectedLetter)
	irregularVerbsPageAsText, err := irregularVerbsManager.GetIrregularVerbsPageAsText(currentPageNumber, selectedLetter)
	if err != nil {
		return fmt.Errorf("failed to get irregular irregularVerbs page as text: %v", err)
	}

	responseText := irregularVerbsPageTitle + irregularVerbsPageAsText

	totalPage, err := irregularVerbsManager.GetTotalPage(selectedLetter)
	if err != nil {
		return fmt.Errorf("failed to get total page: %v", err)
	}
	messageToUser := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, responseText)
	err = irregularVerbsManager.CreateInlineKeyboard(&messageToUser, keyboardVerbValue.Page, totalPage, selectedLetter)
	if err != nil {
		return fmt.Errorf("failed to inline keyboard: %v", err)
	}

	_, err = bot.Send(&messageToUser)
	if err != nil {
		return fmt.Errorf("failed to send response message: %v", err)
	}
	return nil
}
