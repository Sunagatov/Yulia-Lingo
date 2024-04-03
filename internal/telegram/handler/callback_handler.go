package handler

import (
	irregularVerbsCallbackHandler "Yulia-Lingo/internal/irregular_verbs"
	myWordListCallbackHandler "Yulia-Lingo/internal/my_word_list"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

func HandleCallbackQuery(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update) error {
	callbackQuery := botUpdate.CallbackQuery
	callbackMessageFromUser := callbackQuery.Data

	switch {
	case strings.Contains(callbackMessageFromUser, "IrregularVerbs"):
		return irregularVerbsCallbackHandler.HandleIrregularVerbListCallback(callbackQuery, bot)
	case strings.Contains(callbackMessageFromUser, "MyWordList"):
		return myWordListCallbackHandler.HandleIrregularVerbListCallback(callbackQuery, bot)
	default:
		return nil
	}
}
