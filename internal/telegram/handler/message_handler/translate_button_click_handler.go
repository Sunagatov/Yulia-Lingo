package message_handler

import (
	"Yulia-Lingo/internal/api"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"regexp"
)

func HandleDefaultCaseUserMessage(bot *tgbotapi.BotAPI, textFromUser string, chatID int64) error {
	if !isValidWord(textFromUser) {
		responseMessage := "Пожалуйста, отправьте корректное слово на английском языке"
		messageToUser := tgbotapi.NewMessage(chatID, responseMessage)
		_, err := bot.Send(&messageToUser)
		if err != nil {
			return fmt.Errorf("failed to send response message: %v", err)
		}
		return fmt.Errorf("failed to handle default case for user's message because user's message has incorrect format of text")
	}

	responseMessage, err := api.RequestTranslateAPI(textFromUser)
	if err != nil {
		return fmt.Errorf("failed to fetch from API: %v", err)
	}

	messageToUser := tgbotapi.NewMessage(chatID, responseMessage)
	messageToUser.ParseMode = "Markdown"
	messageToUser.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Сохранить", "save_word_option"),
		),
	)
	_, err = bot.Send(&messageToUser)
	if err != nil {
		return fmt.Errorf("failed to send response message: %v", err)
	}
	return nil
}

func isValidWord(word string) bool {
	const pattern = `^[A-Za-z]+$`
	matched, _ := regexp.MatchString(pattern, word)
	return matched
}
