package callback_handler

import (
	irregularVerbsManager "Yulia-Lingo/internal/database/irregular_verbs"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleIrregularVerbListCallback(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) error {
	callbackData := callbackQuery.Data
	callbackChatId := callbackQuery.Message.Chat.ID

	keyboardVerbValue, err := irregularVerbsManager.KeyboardVerbValueFromJSON(callbackData)
	if err != nil {
		return fmt.Errorf("failed to map keyboardVerbValue: %v", err)
	}
	selectedLetter := keyboardVerbValue.Latter
	currentPageNumber := keyboardVerbValue.Page

	irregularVerbsPageAsText, err := irregularVerbsManager.GetIrregularVerbsPageAsText(currentPageNumber, selectedLetter)
	if err != nil {
		return fmt.Errorf("failed to get irregular irregularVerbs page as text: %v", err)
	}

	var responseText string
	if irregularVerbsPageAsText != "" {
		responseText = fmt.Sprintf("Список неправильных глаголов на букву '%s':\n\n", selectedLetter) + irregularVerbsPageAsText
	} else {
		responseText = fmt.Sprintf("Список неправильных глаголов на букву '%s' пуст", selectedLetter)
	}

	keyboard, err := irregularVerbsManager.CreateInlineKeyboard(keyboardVerbValue.Page, selectedLetter)
	if err != nil {
		return fmt.Errorf("failed to inline keyboard: %v", err)
	}

	messageToUser := tgbotapi.NewMessage(callbackChatId, responseText)
	if keyboard != nil {
		messageToUser.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard)
	}

	_, err = bot.Send(&messageToUser)
	if err != nil {
		return fmt.Errorf("failed to send response message: %v", err)
	}
	return nil
}
