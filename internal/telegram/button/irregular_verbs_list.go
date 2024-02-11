package button

import (
	database "Yulia-Lingo/internal/db"
	"Yulia-Lingo/internal/verb/model"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
)

const (
	IrregularVerbsPerPage = 10
)

var userContext = make(map[int64]int)

func GetTotalIrregularVerbsCount(letter string) (int, error) {
	db, err := database.GetPostgresClient()
	if err != nil {
		return 0, fmt.Errorf("can't connect to postgres, err: %v", err)
	}

	var count int
	query := "SELECT COUNT(*) FROM irregular_verbs WHERE verb LIKE $1 || '%'"
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("error preparing statement: %v", err)
	}

	err = stmt.QueryRow(strings.ToLower(letter)).Scan(&count) // Add '%' around letter for LIKE clause
	if err != nil {
		return 0, fmt.Errorf("error getting total irregular verbs count: %v", err)
	}
	defer stmt.Close() // Add this line to close the prepared statement
	return count, nil
}

func GetIrregularVerbs(offset, limit int, letter string) ([]model.IrregularVerb, error) {
	db, err := database.GetPostgresClient()
	if err != nil {
		return nil, fmt.Errorf("can't connect to postgres, err: %v", err)
	}

	query := "SELECT id, original, verb, past, past_participle FROM irregular_verbs WHERE verb LIKE $3 || '%' LIMIT $1 OFFSET $2"
	rows, err := db.Query(query, limit, offset, strings.ToLower(letter))
	if err != nil {
		return nil, fmt.Errorf("error executing database query: %v", err)
	}
	defer rows.Close()

	var verbs []model.IrregularVerb

	for rows.Next() {
		var verb model.IrregularVerb
		err := rows.Scan(&verb.ID, &verb.Original, &verb.Verb, &verb.Past, &verb.PastP)
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

func CreateInlineKeyboard(currentPage int, totalPages int, totalVerbs int, selectedLetter string) tgbotapi.InlineKeyboardMarkup {
	var keyboard []tgbotapi.InlineKeyboardButton
	if currentPage > 1 {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardButtonData("⬅️Назад", GetPaginationCallbackData(currentPage-1, selectedLetter)))
	}
	if currentPage < totalPages && totalVerbs > IrregularVerbsPerPage {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardButtonData("Вперед ➡️", GetPaginationCallbackData(currentPage+1, selectedLetter)))
	}
	return tgbotapi.NewInlineKeyboardMarkup(keyboard)
}

func GetPaginationCallbackData(pageNumber int, selectedLetter string) string {
	return "irregular_verbs_page_" + strconv.Itoa(pageNumber) + "_" + selectedLetter
}

func GetCurrentPage(chatID int64) int {
	if page, ok := userContext[chatID]; ok {
		return page
	}
	return 1
}

func ExtractPageNumber(callbackData string) (int, string) {
	parts := strings.Split(callbackData, "_")
	if len(parts) == 5 && parts[0] == "irregular" && parts[1] == "verbs" && parts[2] == "page" {
		pageNumber, _ := strconv.Atoi(parts[3])
		letter := parts[4]
		return pageNumber, letter
	}
	return 0, ""
}

func UpdateCurrentPage(chatID int64, pageNumber int) {
	userContext[chatID] = pageNumber
}
