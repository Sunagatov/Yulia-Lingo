package bot

import (
	"Yulia-Lingo/internal/i18n"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const parseMode = "Markdown"

func BuildMainKeyboard(msgSource *i18n.MessageSource, lang i18n.Lang) tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(msgSource.Get(lang, i18n.MsgLabelIrregularVerbs)),
			tgbotapi.NewKeyboardButton(msgSource.Get(lang, i18n.MsgLabelMyWordList)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(msgSource.Get(lang, i18n.MsgLabelLang)),
			tgbotapi.NewKeyboardButton(msgSource.Get(lang, i18n.MsgLabelMenu)),
		),
	)
	kb.ResizeKeyboard = true
	return kb
}

func NewMessage(chatID int64, text string) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = parseMode
	return msg
}

func NewMessageWithKeyboard(chatID int64, text string, keyboard interface{}) tgbotapi.MessageConfig {
	msg := NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	return msg
}

func NewEditMessage(chatID int64, messageID int, text string) tgbotapi.EditMessageTextConfig {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = parseMode
	return msg
}

func NewEditMessageWithKeyboard(chatID int64, messageID int, text string, keyboard *tgbotapi.InlineKeyboardMarkup) tgbotapi.EditMessageTextConfig {
	msg := NewEditMessage(chatID, messageID, text)
	msg.ReplyMarkup = keyboard
	return msg
}
