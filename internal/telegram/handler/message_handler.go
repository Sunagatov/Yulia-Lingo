package handler

import (
	"Yulia-Lingo/internal/irregular_verbs"
	"Yulia-Lingo/internal/my_word_list"
	"Yulia-Lingo/internal/translate"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	StartTelegramBotCommand = "/start"
	IrregularVerbsCommand   = "🔺 Неправильные глаголы"
	MyWordListCommand       = "🔺 Мой список слов"
)

func HandleMessageFromUser(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update) error {
	chatID := botUpdate.Message.Chat.ID
	messageFromUser := botUpdate.Message.Text

	switch messageFromUser {
	case StartTelegramBotCommand:
		return HandleStartButtonClick(bot, botUpdate, chatID)
	case IrregularVerbsCommand:
		return irregular_verbs.HandleIrregularVerbsButtonClick(bot, chatID)
	case MyWordListCommand:
		return my_word_list.HandleMyWordButtonClick(bot, chatID)
	default:
		return translate.HandleDefaultCaseUserMessage(bot, messageFromUser, chatID)
	}
}
