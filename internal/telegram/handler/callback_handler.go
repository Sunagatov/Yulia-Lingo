package handler

import (
	irregularVerbsManager "Yulia-Lingo/internal/database/irregular_verbs"
	"fmt"
	tgBotApi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

func HandleCallbackQuery(bot *tgBotApi.BotAPI, botUpdate tgBotApi.Update) error {
	callbackQuery := botUpdate.CallbackQuery
	callbackMessageFromUser := callbackQuery.Data

	switch {
	case strings.HasPrefix(callbackMessageFromUser, "select_letter_"):
		return handleSelectLetterCallback(bot, callbackQuery)
	case strings.HasPrefix(callbackMessageFromUser, "irregular_verbs_page_"):
		return handleIrregularVerbsPaginationButtonsCallback(bot, callbackQuery)
	default:
		return nil
	}
}

func handleIrregularVerbsPaginationButtonsCallback(bot *tgBotApi.BotAPI, callbackQuery *tgBotApi.CallbackQuery) error {
	callbackChatID := callbackQuery.Message.Chat.ID
	callbackMessageID := callbackQuery.Message.MessageID
	callbackMessageText := callbackQuery.Message.Text
	callbackMessageFromUser := callbackQuery.Data

	pageNumber, selectedLetter := irregularVerbsManager.ExtractPageNumber(callbackMessageFromUser)

	irregularVerbsManager.UpdateCurrentPage(callbackChatID, pageNumber)

	msg := tgBotApi.NewEditMessageText(callbackChatID, callbackMessageID, callbackMessageText)
	_, err := bot.Send(&msg)
	if err != nil {
		return fmt.Errorf("failed to send the message to a user: %v", err)
	}

	totalVerbs, err := irregularVerbsManager.GetTotalIrregularVerbsCount(selectedLetter)
	if err != nil {
		return fmt.Errorf("failed to get the total irregularVerbs count: %v", err)
	}

	if totalVerbs == 0 {
		responseText := fmt.Sprintf("К сожалению глаголов начинающихся на латинскую букву - '%s' нет.\n\n", selectedLetter)
		messageToUser := tgBotApi.NewMessage(callbackChatID, responseText)
		_, err = bot.Send(&messageToUser)
		if err != nil {
			return fmt.Errorf("failed to send the message to a user: %v", err)
		}
		return nil
	} else {
		currentPageNumber, err := irregularVerbsManager.GetCurrentPageNumber(callbackChatID)
		if err != nil {
			return err
		}

		irregularVerbsPageAsText, err := irregularVerbsManager.GetIrregularVerbsPageAsText(currentPageNumber, selectedLetter)
		if err != nil {
			return fmt.Errorf("failed to get irregular irregularVerbs page as text: %v", err)
		}

		messageToUser := tgBotApi.NewMessage(callbackChatID, irregularVerbsPageAsText)
		if totalVerbs > irregularVerbsManager.IrregularVerbsCountPerPage {
			messageToUser.ReplyMarkup = irregularVerbsManager.CreateInlinePaginationButtonsForIrregularVerbsPage(currentPageNumber, totalVerbs, selectedLetter)
		}

		_, errorMessage := bot.Send(&messageToUser)
		if errorMessage != nil {
			return fmt.Errorf("failed to send response message: %v", errorMessage)
		}
		return nil
	}
}

func handleSelectLetterCallback(bot *tgBotApi.BotAPI, callbackQuery *tgBotApi.CallbackQuery) error {
	callbackChatID := callbackQuery.Message.Chat.ID
	callbackMessageFromUser := callbackQuery.Data
	selectedLetter := strings.TrimPrefix(callbackMessageFromUser, "select_letter_")

	pageNumber := 1
	irregularVerbsManager.UpdateCurrentPage(callbackChatID, pageNumber)

	totalVerbs, err := irregularVerbsManager.GetTotalIrregularVerbsCount(selectedLetter)
	if err != nil {
		return fmt.Errorf("failed to get total irregular irregularVerbs count: %v", err)
	}

	if totalVerbs == 0 {
		textOfMessageToUser := fmt.Sprintf("К сожалению глаголов начинающихся на латинскую букву - '%s' нет.\n\n", selectedLetter)
		messageToUser := tgBotApi.NewMessage(callbackChatID, textOfMessageToUser)
		_, err = bot.Send(&messageToUser)
		if err != nil {
			return fmt.Errorf("failed to send response message: %v", err)
		}
		return nil
	} else {
		currentPageNumber, err := irregularVerbsManager.GetCurrentPageNumber(callbackChatID)
		if err != nil {
			return err
		}

		irregularVerbsPageAsText, err := irregularVerbsManager.GetIrregularVerbsPageAsText(currentPageNumber, selectedLetter)
		if err != nil {
			return fmt.Errorf("failed to get irregular irregularVerbs page as text: %v", err)
		}

		titleInTextMessageToUser := fmt.Sprintf("Список неправильных глаголов начинающихся на латинскую букву - '%s':\n\n", selectedLetter)
		textOfMessageToUser := titleInTextMessageToUser + irregularVerbsPageAsText

		messageToUser := tgBotApi.NewMessage(callbackChatID, textOfMessageToUser)
		if totalVerbs > irregularVerbsManager.IrregularVerbsCountPerPage {
			messageToUser.ReplyMarkup = irregularVerbsManager.CreateInlinePaginationButtonsForIrregularVerbsPage(currentPageNumber, totalVerbs, selectedLetter)
		}
		_, errorMessage := bot.Send(&messageToUser)
		if errorMessage != nil {
			return fmt.Errorf("failed to send response message: %v", errorMessage)
		}
		return nil
	}
}
