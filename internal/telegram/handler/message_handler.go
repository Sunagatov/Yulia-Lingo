package handler

import (
	messageHandler "Yulia-Lingo/internal/telegram/handler/message_handler"
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
		return messageHandler.HandleStartButtonClick(bot, botUpdate, chatID)
	case IrregularVerbsCommand:
		return messageHandler.HandleIrregularVerbsButtonClick(bot, chatID)
	default:
		return nil
	}
}
