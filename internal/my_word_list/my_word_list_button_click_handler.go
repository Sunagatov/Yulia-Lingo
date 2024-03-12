package my_word_list

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleMyWordButtonClick(bot *tgbotapi.BotAPI, chatID int64) error {
	messageText := "*Ваш список слов пуст*\n\n"
	messageToUser := tgbotapi.NewMessage(chatID, messageText)
	messageToUser.ParseMode = "Markdown"

	_, err := bot.Send(&messageToUser)
	if err != nil {
		return fmt.Errorf("failed to send the message for 'IrregularVerbs' button to a user: %v", err)
	}
	return nil
}
