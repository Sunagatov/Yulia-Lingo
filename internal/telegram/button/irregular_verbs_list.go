package button

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"math"
	"strconv"
	"strings"
)

var irregularVerbs = []string{
	"Идти - [Go - Went - Gone]",
	"Петь - [Sing - Sang - Sung]",
	"Есть - [Eat - Ate - Eaten]",
	"Спать - [Sleep - Slept - Slept]",
	"Говорить - [Speak - Spoke - Spoken]",
	"Брать - [Take - Took - Taken]",
	"Бежать - [Run - Ran - Run]",
	"Читать - [Read - Read - Read]",
	"Писать - [Write - Wrote - Written]",
	"Плавать - [Swim - Swam - Swum]",
	"Лететь - [Fly - Flew - Flown]",
	"Водить - [Drive - Drove - Driven]",
	"Ломать - [Break - Broke - Broken]",
	"Строить - [Build - Built - Built]",
	"Выбирать - [Choose - Chose - Chosen]",
	"Забывать - [Forget - Forgot - Forgotten]",
	"Встречать - [Meet - Met - Met]",
	"Думать - [Think - Thought - Thought]",
	"Учить - [Teach - Taught - Taught]",
	"Видеть - [See - Saw - Seen]",
	"Пить - [Drink - Drank - Drunk]",
	"Иметь - [Have - Had - Had]",
	"Делать - [Do - Did - Done]",
	"Говорить - [Say - Said - Said]",
	"Покупать - [Buy - Bought - Bought]",
	"Ломать - [Break - Broke - Broken]",
	"Начинать - [Begin - Began - Begun]",
	"Выбирать - [Choose - Chose - Chosen]",
	"Падать - [Fall - Fell - Fallen]",
	"Знать - [Know - Knew - Known]",
	"Говорить - [Speak - Spoke - Spoken]",
	"Спать - [Sleep - Slept - Slept]",
	"Находить - [Find - Found - Found]",
	"Терять - [Lose - Lost - Lost]",
	"Выигрывать - [Win - Won - Won]",
	"Рисовать - [Draw - Drew - Drawn]",
	"Держать - [Hold - Held - Held]",
	"Делать - [Make - Made - Made]",
	"Платить - [Pay - Paid - Paid]",
}

const (
	IrregularVerbsPerPage = 10
)

var userContext = make(map[int64]int)

func HandleIrregularVerbsListButtonClick(bot *tgbotapi.BotAPI, chatID int64) {
	// Get the current page number from the user's context
	currentPage := getCurrentPage(chatID)

	// Calculate the total number of pages
	totalPages := int(math.Ceil(float64(len(irregularVerbs)) / IrregularVerbsPerPage))

	// Get the current page's verbs
	startIndex := (currentPage - 1) * IrregularVerbsPerPage
	endIndex := startIndex + IrregularVerbsPerPage
	currentPageVerbs := irregularVerbs[startIndex:min(endIndex, len(irregularVerbs))]

	// Create the message text with the current page's verbs
	messageText := strings.Join(currentPageVerbs, "\n")

	// Send the message to the user
	messageToUser := tgbotapi.NewMessage(chatID, messageText)
	messageToUser.ReplyMarkup = createInlineKeyboard(currentPage, totalPages)

	_, errorMessage := bot.Send(messageToUser)
	if errorMessage != nil {
		log.Printf("Error sending response message: %v", errorMessage)
	}
}

func createInlineKeyboard(currentPage, totalPages int) tgbotapi.InlineKeyboardMarkup {
	var rows []tgbotapi.InlineKeyboardButton
	if currentPage > 1 {
		rows = append(rows, tgbotapi.NewInlineKeyboardButtonData("Prev page", getPaginationCallbackData(currentPage-1)))
	}
	if currentPage < totalPages {
		rows = append(rows, tgbotapi.NewInlineKeyboardButtonData("Next page", getPaginationCallbackData(currentPage+1)))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows)
}

// Function to get the callback data for pagination
func getPaginationCallbackData(pageNumber int) string {
	return "irregular_verbs_page_" + strconv.Itoa(pageNumber)
}

// Function to get the current page number from user's context
func getCurrentPage(chatID int64) int {
	// You may use a database or another storage mechanism to store user's context
	// For simplicity, a map is used here
	if page, ok := userContext[chatID]; ok {
		return page
	}
	return 1
}

// Function to calculate the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ExtractPageNumber(callbackData string) int {
	parts := strings.Split(callbackData, "_")
	if len(parts) == 4 && parts[0] == "irregular" && parts[1] == "verbs" && parts[2] == "page" {
		pageNumber, _ := strconv.Atoi(parts[3])
		return pageNumber
	}
	return 0
}

func UpdateCurrentPage(chatID int64, pageNumber int) {
	// You may use a database or another storage mechanism to store user's context
	// For simplicity, a map is used here
	userContext[chatID] = pageNumber
}
