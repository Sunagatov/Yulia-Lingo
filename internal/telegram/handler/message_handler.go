package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
)

func HandleMessageFromUser(bot *tgbotapi.BotAPI, botUpdate tgbotapi.Update) {
	messageFromUser := botUpdate.Message
	chatID := messageFromUser.Chat.ID
	textFromUser := messageFromUser.Text

	switch textFromUser {
	case "/start":
		{
			userName := os.Getenv("TG_BOT_NAME")
			greetingMessageToUser := "Привет, %s!\nЭто телеграм бот - Yulia-lingo.\nОн поможет тебе пополнить твой словарный запас по английскому языку."
			text := fmt.Sprintf(greetingMessageToUser, userName)
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("🔺 Неправильные глаголы")),
			)
			_, errorMessage := bot.Send(&msg)

			if errorMessage != nil {
				log.Printf("Error sending response message: %v", errorMessage)
			}

		}
	case "🔺 Неправильные глаголы":
		{
			keyboard := CreateLetterKeyboardMarkup()

			messageText := "С какой буквы вы хотите начать изучение неправильных глаголов?"

			message := tgbotapi.NewMessage(chatID, messageText)
			message.ReplyMarkup = keyboard

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
