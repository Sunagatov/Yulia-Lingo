package model

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Message struct {
	TelegramBot *tgbotapi.BotAPI
	ChatId      int64            `json:"chat_id"`
	Text        string           `json:"text"`
	Keyboard    []KeyboardButton `json:"reply_markup"`
}

type KeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}
