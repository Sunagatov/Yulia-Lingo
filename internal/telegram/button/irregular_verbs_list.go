package button

import (
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"math"
	"strconv"
	"strings"
)

const (
	IrregularVerbsPerPage = 10
)

type IrregularVerb struct {
	ID    int    `json:"id"`
	Verb  string `json:"verb"`
	Past  string `json:"past"`
	PastP string `json:"past_participle"`
}

var userContext = make(map[int64]int)

func HandleIrregularVerbsListButtonClick(bot *tgbotapi.BotAPI, db *sql.DB, chatID int64) {
	// Get the current page number from the user's context
	currentPage := getCurrentPage(chatID)

	// Calculate the total number of pages
	totalVerbs, err := getTotalIrregularVerbsCount(db)
	if err != nil {
		log.Printf("Error getting total irregular verbs count: %v", err)
		return
	}
	totalPages := int(math.Ceil(float64(totalVerbs) / IrregularVerbsPerPage))

	// Get the current page's verbs from the database
	offset := (currentPage - 1) * IrregularVerbsPerPage
	verbs, err := getIrregularVerbs(db, offset, IrregularVerbsPerPage)
	if err != nil {
		log.Printf("Error getting irregular verbs: %v", err)
		return
	}

	// Create the message text with the current page's verbs
	var messageText string
	for _, verb := range verbs {
		messageText += fmt.Sprintf("%s - [%s - %s - %s]\n", verb.Verb, verb.Past, verb.PastP)
	}

	// Send the message to the user
	messageToUser := tgbotapi.NewMessage(chatID, messageText)
	messageToUser.ReplyMarkup = createInlineKeyboard(currentPage, totalPages)

	_, errorMessage := bot.Send(messageToUser)
	if errorMessage != nil {
		log.Printf("Error sending response message: %v", errorMessage)
	}
}

func getTotalIrregularVerbsCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM irregular_verbs").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting total irregular verbs count: %v", err)
	}
	return count, nil
}

func getIrregularVerbs(db *sql.DB, offset, limit int) ([]IrregularVerb, error) {
	query := "SELECT id, verb, past, past_participle FROM irregular_verbs ORDER BY id LIMIT $1 OFFSET $2"
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error executing database query: %v", err)
	}
	defer rows.Close()

	var verbs []IrregularVerb

	for rows.Next() {
		var verb IrregularVerb
		err := rows.Scan(&verb.ID, &verb.Verb, &verb.Past, &verb.PastP)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		verbs = append(verbs, verb)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	return verbs, nil
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
