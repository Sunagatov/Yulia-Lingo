package handler

import (
	callbackHandler "Yulia-Lingo/internal/telegram/handler/callback_handler"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

func HandleCallbackQuery(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update) error {
	callbackQuery := botUpdate.CallbackQuery
	callbackMessageFromUser := callbackQuery.Data

	switch {
	case strings.Contains(callbackMessageFromUser, "IrregularVerbs"):
		return callbackHandler.HandleIrregularVerbListCallback(callbackQuery, bot)
	default:
		return nil
	}
}
