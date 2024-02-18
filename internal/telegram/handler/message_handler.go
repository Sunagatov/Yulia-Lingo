package handler

import (
	"Yulia-Lingo/internal/irregular_verbs"
	"Yulia-Lingo/internal/translate"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	StartTelegramBotCommand = "/start"
	IrregularVerbsCommand   = "🔺 Неправильные глаголы"
)

func HandleMessageFromUser(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update) error {
	chatID := botUpdate.Message.Chat.ID
	messageFromUser := botUpdate.Message.Text

	switch messageFromUser {
	case StartTelegramBotCommand:
		return HandleStartButtonClick(bot, botUpdate, chatID)
	case IrregularVerbsCommand:
		return irregular_verbs.HandleIrregularVerbsButtonClick(bot, chatID)
	default:
		return translate.HandleDefaultCaseUserMessage(bot, messageFromUser, chatID)
	}
}
