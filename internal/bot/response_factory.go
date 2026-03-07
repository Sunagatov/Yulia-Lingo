package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const parseMode = "Markdown"

type ResponseFactory struct{}

func (ResponseFactory) NewTextMessage(chatID int64, text string) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = parseMode
	return msg
}

func (ResponseFactory) NewTextMessageWithKeyboard(chatID int64, text string, keyboard interface{}) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = parseMode
	msg.ReplyMarkup = keyboard
	return msg
}

func (ResponseFactory) NewEditMessage(chatID int64, messageID int, text string) tgbotapi.EditMessageTextConfig {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = parseMode
	return msg
}

func (ResponseFactory) NewEditMessageWithKeyboard(chatID int64, messageID int, text string, keyboard *tgbotapi.InlineKeyboardMarkup) tgbotapi.EditMessageTextConfig {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = parseMode
	msg.ReplyMarkup = keyboard
	return msg
}
