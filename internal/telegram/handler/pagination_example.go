package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log" // Import the log package
	"strconv"
)

const (
	PrevCallbackData = "prev"
	NextCallbackData = "next"
)

var itemsPerPage = 5

const (
	pageSize = 5
)

// UserPageMap stores the current page for each user
var UserPageMap = make(map[int64]int)

func handleBotUpdates(bot *tgbotapi.BotAPI) {
	updateEndpoint := "/" + bot.Token

	botUpdates := bot.ListenForWebhook(updateEndpoint)

	for update := range botUpdates {
		if update.Message != nil {
			switch update.Message.Text {
			case "/start":
				sendWelcomeMessage(bot, update.Message.Chat.ID)
				handlePagination(bot, update.Message.Chat.ID, 1)
			}
		}

		if update.CallbackQuery != nil {
			callbackData := update.CallbackQuery.Data
			currentPage, err := strconv.Atoi(callbackData)
			if err == nil {
				log.Printf("Callback received. CallbackData: %s, CurrentPage: %d", callbackData, currentPage)
				switch callbackData {
				case PrevCallbackData:
					log.Printf("Handling pagination for Previous button")
					handlePagination(bot, update.CallbackQuery.Message.Chat.ID, currentPage-1)
				case NextCallbackData:
					log.Printf("Handling pagination for Next button")
					handlePagination(bot, update.CallbackQuery.Message.Chat.ID, currentPage+1)
				default:
					log.Printf("Handling pagination for other cases")
					handlePagination(bot, update.CallbackQuery.Message.Chat.ID, currentPage)
				}
			} else {

			}
		}
	}
}

func sendWelcomeMessage(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Welcome! Type /start to start pagination.")
	bot.Send(msg)
}

func handlePagination(bot *tgbotapi.BotAPI, chatID int64, currentPage int) {
	// Ensure currentPage is not less than 1
	if currentPage < 1 {
		currentPage = 1
	}

	startIndex := (currentPage - 1) * pageSize
	endIndex := currentPage * pageSize

	// Assuming you have some data to paginate, for example, a list of strings.
	data := generateData()

	// Send a message with the current page items
	msgText := "Page " + strconv.Itoa(currentPage) + ":\n"
	for i := startIndex; i < endIndex && i < len(data); i++ {
		msgText += data[i] + "\n"
	}

	msg := tgbotapi.NewMessage(chatID, msgText)

	// Create inline keyboard for pagination
	inlineKeyboard := &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.NewInlineKeyboardButtonData("Previous", PrevCallbackData),
				tgbotapi.NewInlineKeyboardButtonData("Next", NextCallbackData),
			},
		},
	}

	msg.ReplyMarkup = inlineKeyboard

	// If the user has already interacted, edit the existing message
	if messageID, exists := UserPageMap[chatID]; exists {
		log.Printf("Editing existing message. ChatID: %d, MessageID: %d", chatID, messageID)
		editMsg := tgbotapi.NewEditMessageText(chatID, messageID, msgText+" (Updated)")
		editMsg.ReplyMarkup = inlineKeyboard
		bot.Send(editMsg)
	} else {
		// If it's the first interaction, send a new message
		log.Printf("Sending new message. ChatID: %d, CurrentPage: %d", chatID, currentPage)
		sentMsg, err := bot.Send(msg)
		if err == nil {
			// Save the message ID and current page for future interactions
			UserPageMap[chatID] = sentMsg.MessageID
			log.Printf("Message sent successfully. ChatID: %d, MessageID: %d", chatID, sentMsg.MessageID)
		} else {
			log.Printf("Error sending message: %v", err)
		}
	}
}

func generateData() []string {
	var data []string
	for i := 1; i <= 50; i++ {
		data = append(data, fmt.Sprintf("Item %d", i))
	}
	return data
}
