package message

import (
	"Yulia-Lingo/internal/telegram/setup"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Send(message *tgbotapi.MessageConfig) error {
	bot, err := setup.GetTelegramBot()
	if err != nil {
		return fmt.Errorf("app wosn't connect to telegram bot, err: %v", err)
	}

	_, err = bot.Send(message)
	return err
}

func Edit(message *tgbotapi.EditMessageTextConfig) error {
	bot, err := setup.GetTelegramBot()
	if err != nil {
		return fmt.Errorf("app wosn't connect to telegram bot, err: %v", err)
	}

	_, err = bot.Send(message)
	return err
}
