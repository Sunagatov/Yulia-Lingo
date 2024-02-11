package handler

import (
	"Yulia-Lingo/internal/telegram/button"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

func HandleCallbackQuery(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update) error {
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

func handleIrregularVerbsPaginationButtonsCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) error {
	callbackChatID := callbackQuery.Message.Chat.ID
	callbackMessageID := callbackQuery.Message.MessageID
	callbackMessageText := callbackQuery.Message.Text
	callbackMessageFromUser := callbackQuery.Data

	pageNumber, selectedLetter := button.ExtractPageNumber(callbackMessageFromUser)

	button.UpdateCurrentPage(callbackChatID, pageNumber)

	msg := tgbotapi.NewEditMessageText(callbackChatID, callbackMessageID, callbackMessageText)
	_, err := bot.Send(&msg)
	if err != nil {
		log.Printf("Error with edit message, err: %v", err)
	}

	totalVerbs, err := button.GetTotalIrregularVerbsCount(selectedLetter)
	if err != nil {
		return fmt.Errorf("error getting total irregular irregularVerbs count: %v", err)
	}

	if totalVerbs == 0 {
		responseText := fmt.Sprintf("К сожалению глаголов начинающихся на латинскую букву - '%s' нет.\n\n", selectedLetter)
		messageToUser := tgbotapi.NewMessage(callbackChatID, responseText)
		_, errorMessage := bot.Send(&messageToUser)
		if errorMessage != nil {
			return fmt.Errorf("error sending response message for the%s: %v", errorMessage, callbackMessageFromUser)
		}
		return nil
	} else {
		currentPageNumber, err := button.GetCurrentPageNumber(callbackChatID)
		if err != nil {
			return err
		}

		irregularVerbsPageAsText, err := button.GetIrregularVerbsPageAsText(currentPageNumber, selectedLetter)
		if err != nil {
			return fmt.Errorf("error getting irregular irregularVerbs page as text: %v", err)
		}

		messageToUser := tgbotapi.NewMessage(callbackChatID, irregularVerbsPageAsText)
		if totalVerbs > button.IrregularVerbsCountPerPage {
			messageToUser.ReplyMarkup = button.CreateInlinePaginationButtonsForIrregularVerbsPage(currentPageNumber, totalVerbs, selectedLetter)
		}

		_, errorMessage := bot.Send(&messageToUser)
		if errorMessage != nil {
			return fmt.Errorf("error sending response message: %v", errorMessage)
		}
		return nil
	}
}

func handleSelectLetterCallback(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) error {
	callbackChatID := callbackQuery.Message.Chat.ID
	callbackMessageFromUser := callbackQuery.Data
	selectedLetter := strings.TrimPrefix(callbackMessageFromUser, "select_letter_")

	pageNumber := 1
	button.UpdateCurrentPage(callbackChatID, pageNumber)

	totalVerbs, err := button.GetTotalIrregularVerbsCount(selectedLetter)
	if err != nil {
		return fmt.Errorf("error getting total irregular irregularVerbs count: %v", err)
	}

	if totalVerbs == 0 {
		textOfMessageToUser := fmt.Sprintf("К сожалению глаголов начинающихся на латинскую букву - '%s' нет.\n\n", selectedLetter)
		messageToUser := tgbotapi.NewMessage(callbackChatID, textOfMessageToUser)
		_, err = bot.Send(&messageToUser)
		if err != nil {
			return fmt.Errorf("error sending response message: %v", err)
		}
		return nil
	} else {
		currentPageNumber, err := button.GetCurrentPageNumber(callbackChatID)
		if err != nil {
			return err
		}

		irregularVerbsPageAsText, err := button.GetIrregularVerbsPageAsText(currentPageNumber, selectedLetter)
		if err != nil {
			return fmt.Errorf("error getting irregular irregularVerbs page as text: %v", err)
		}

		titleInTextMessageToUser := fmt.Sprintf("Список неправильных глаголов начинающихся на латинскую букву - '%s':\n\n", selectedLetter)
		textOfMessageToUser := titleInTextMessageToUser + irregularVerbsPageAsText

		messageToUser := tgbotapi.NewMessage(callbackChatID, textOfMessageToUser)
		if totalVerbs > button.IrregularVerbsCountPerPage {
			messageToUser.ReplyMarkup = button.CreateInlinePaginationButtonsForIrregularVerbsPage(currentPageNumber, totalVerbs, selectedLetter)
		}
		_, errorMessage := bot.Send(&messageToUser)
		if errorMessage != nil {
			return fmt.Errorf("error sending response message: %v", errorMessage)
		}
		return nil
	}
}
