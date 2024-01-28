package handler

import (
	"Yulia-Lingo/internal/api"
	tgbutton "Yulia-Lingo/internal/telegram/button"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"regexp"
)

type TgCommandHandler func(bot *tgbotapi.BotAPI, chatID int64)

var tgCommandHandlers = map[string]TgCommandHandler{
	tgbutton.StartButtonName:      handleStartButtonClick,
	tgbutton.MyWordListButtonName: tgbutton.HandleMyWordListButtonClick,
	tgbutton.SaveWordButtonName:   handleSaveWordButtonClick,
	tgbutton.VerbsButtonName:      tgbutton.HandleVerbsButtonClick,
}

func HandleMessageFromUser(bot *tgbotapi.BotAPI, messageFromUser *tgbotapi.Message) {
	textFromUser := messageFromUser.Text
	chatID := messageFromUser.Chat.ID

	tgCommandHandler, ok := tgCommandHandlers[textFromUser]

	if ok {
		tgCommandHandler(bot, chatID)
	} else {
		handleDefaultCaseUserMessage(bot, textFromUser, chatID)
	}
}

func handleSaveWordButtonClick(bot *tgbotapi.BotAPI, chatID int64) {
	responseText := "Слово сохранено для последующего изучения"
	msg := tgbotapi.NewMessage(chatID, responseText)
	_, errorMessage := bot.Send(msg)
	if errorMessage != nil {
		log.Printf("Error sending response message: %v", errorMessage)
	}
}

func handleStartButtonClick(bot *tgbotapi.BotAPI, chatID int64) {
	userName := bot.Self.FirstName
	text := fmt.Sprintf(tgbutton.GreetingMessageToUser, userName)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(tgbutton.MyWordListButtonName)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(tgbutton.IrregularVerbListButtonName)),
	)
	_, errorMessage := bot.Send(msg)
	if errorMessage != nil {
		log.Printf("Error sending response message: %v", errorMessage)
	}
}

func handleDefaultCaseUserMessage(bot *tgbotapi.BotAPI, textFromUser string, chatID int64) {
	if !isValidWord(textFromUser) {
		responseMessage := "Пожалуйста, отправьте корректное слово на английском языке"
		messageToUser := tgbotapi.NewMessage(chatID, responseMessage)
		_, sendingMessageError := bot.Send(messageToUser)
		if sendingMessageError != nil {
			log.Printf("Error sending message to a user: %v", sendingMessageError)
		}
	}

	responseMessage, err := api.RequestTranslateAPI(textFromUser)
	if err != nil {
		log.Printf("Error fetching from API: %v", err)
		responseMessage = "Sorry, there was an error processing your request."
	}
	messageToUser := tgbotapi.NewMessage(chatID, responseMessage)

	messageToUser.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Сохранить", "save_word_option"),
		),
	)
	_, errorMessage := bot.Send(messageToUser)
	if errorMessage != nil {
		log.Printf("Error sending response message: %v", errorMessage)
	}
}

func isValidWord(word string) bool {
	const pattern = `^[A-Za-z]+$`
	matched, _ := regexp.MatchString(pattern, word)
	return matched
}
