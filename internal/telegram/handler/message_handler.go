package handler

import (
	"Yulia-Lingo/internal/verb/model"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func HandleMessageFromUser(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update) {
	messageFromUser := botUpdate.Message
	chatID := messageFromUser.Chat.ID
	textFromUser := messageFromUser.Text

	switch textFromUser {
	case "/start":
		{
			userName := botUpdate.Message.From.UserName

			greetingMessageToUser := "Здравствуйте, %s!\n\nЭто телеграм бот - Yulia-lingo.\n\nБот поможет тебе пополнить словарный запас английского языка.\n\nСейчас доступен:\n- Список неправильных глаголов."
			text := fmt.Sprintf(greetingMessageToUser, userName)
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("🔺 Неправильные глаголы")),
			)
			_, errorMessage := bot.Send(&msg)

			if errorMessage != nil {
				log.Printf("Error sending response message for /start: %v", errorMessage)
			}

		}
	case "🔺 Неправильные глаголы":
		{
			messageText := "С какой буквы вы хотите начать изучение неправильных глаголов?\n------------------------\n"

			message := tgbotapi.NewMessage(chatID, messageText)
			message.ReplyMarkup = CreateLetterKeyboardMarkup()

			_, err := bot.Send(&message)
			if err != nil {
				log.Printf("Error sending message: %v", err)
			}
		}
	default:
	}
}

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func CreateLetterKeyboardMarkup() tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	var currentRow []tgbotapi.InlineKeyboardButton

	for _, letter := range letters {
		latterStr := string(letter)

		btn := tgbotapi.NewInlineKeyboardButtonData(latterStr, model.KeyboardVerbValue{
			Request: "GetListByLatter",
			Page:    0,
			Latter:  latterStr,
		}.ToJSON())

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
