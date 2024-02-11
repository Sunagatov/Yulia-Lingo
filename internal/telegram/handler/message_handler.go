package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	letters            = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	greetingBotMessage = "Здравствуйте, %s %s!\n\nЭто телеграм бот - Yulia-lingo.\n\n" +
		"Бот поможет вам пополнить словарный запас английского языка.\n\n" +
		"Сейчас доступен:\n- Список неправильных глаголов."
)

func HandleMessageFromUser(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update) error {
	chatID := botUpdate.Message.Chat.ID
	messageFromUser := botUpdate.Message.Text

	switch messageFromUser {
	case "/start":
		{
			userFirstName := botUpdate.Message.From.FirstName
			userLastName := botUpdate.Message.From.LastName
			greetingMessage := fmt.Sprintf(greetingBotMessage, userFirstName, userLastName)
			messageToUser := tgbotapi.NewMessage(chatID, greetingMessage)
			messageToUser.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("🔺 Неправильные глаголы")),
			)
			_, errorMessage := bot.Send(&messageToUser)
			if errorMessage != nil {
				return fmt.Errorf("error sending greeting message to as user: %v", errorMessage)
			}
		}
	case "🔺 Неправильные глаголы":
		{
			messageText := "С какой буквы вы хотите начать изучение неправильных глаголов?\n\n"
			messageToUser := tgbotapi.NewMessage(chatID, messageText)
			messageToUser.ReplyMarkup = CreateLetterKeyboardMarkup()
			_, err := bot.Send(&messageToUser)
			if err != nil {
				return fmt.Errorf("error sending messageToUser for 'IrregularVerbs' button: %v", err)
			}
		}
	default:
	}
	return nil
}

func CreateLetterKeyboardMarkup() tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for _, letter := range letters {
		btn := tgbotapi.NewInlineKeyboardButtonData(string(letter), "select_letter_"+string(letter))
		currentRow = append(currentRow, btn)
		if len(currentRow) == 5 {
			rows = append(rows, currentRow)
			currentRow = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
