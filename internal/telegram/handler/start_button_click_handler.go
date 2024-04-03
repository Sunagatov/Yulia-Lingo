package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	greetingBotMessage = "Здравствуйте, *%s %s*!\n\nЭто телеграм бот - *Yulia-lingo*.\n\n" +
		"Бот поможет вам пополнить *словарный запас английского языка*.\n\n" +
		"*Сейчас доступен:*\n- Список неправильных глаголов."
)

func HandleStartButtonClick(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update, chatID int64) error {
	{
		userFirstName := botUpdate.Message.From.FirstName
		userLastName := botUpdate.Message.From.LastName
		greetingMessage := fmt.Sprintf(greetingBotMessage, userFirstName, userLastName)
		messageToUser := tgbotapi.NewMessage(chatID, greetingMessage)
		messageToUser.ParseMode = "Markdown"
		messageToUser.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("🔺 Неправильные глаголы")),
			tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("🔺 Мой список слов")),
		)
		_, errorMessage := bot.Send(&messageToUser)
		if errorMessage != nil {
			return fmt.Errorf("failed to send the greeting message to a user: %v", errorMessage)
		}
	}
	return nil
}
