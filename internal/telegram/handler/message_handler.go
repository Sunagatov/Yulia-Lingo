package handler

import (
	tgbutton "Yulia-Lingo/internal/telegram/button"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleMessageFromUser(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update) {
	messageFromUser := botUpdate.Message
	chatID := messageFromUser.Chat.ID
	textFromUser := messageFromUser.Text

	switch textFromUser {
	case tgbutton.StartButtonName:
		tgbutton.HandleStartButtonClick(bot, chatID)
	case tgbutton.IrregularVerbListButtonName:
		tgbutton.HandleIrregularVerbsListButtonClick(bot, chatID)
	default:
	}
}
