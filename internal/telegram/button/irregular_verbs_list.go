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

func GetTotalIrregularVerbsCount() (int, error) {
	db, err := database.GetPostgresClient()
	if err != nil {
		return 0, fmt.Errorf("can't connect to postgres, err: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM irregular_verbs").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting total irregular verbs count: %v", err)
	}
	return count, err
}

func GetIrregularVerbs(offset, limit int) ([]model.IrregularVerb, error) {
	db, err := database.GetPostgresClient()
	if err != nil {
		return nil, fmt.Errorf("can't connect to postgres, err: %v", err)
	}

	query := "SELECT id, translated, verb, past, past_participle FROM irregular_verbs LIMIT $1 OFFSET $2"
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error executing database query: %v", err)
	}
	defer rows.Close()

	var verbs []model.IrregularVerb

	for rows.Next() {
		var verb model.IrregularVerb
		err := rows.Scan(&verb.ID, &verb.Verb, &verb.Original, &verb.Past, &verb.PastP)
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

func CreateInlineKeyboard(currentPage, totalPages int) tgbotapi.InlineKeyboardMarkup {
	var rows []tgbotapi.InlineKeyboardButton
	if currentPage > 1 {
		rows = append(rows, tgbotapi.NewInlineKeyboardButtonData("Prev page", GetPaginationCallbackData(currentPage-1)))
	}
	if currentPage < totalPages {
		rows = append(rows, tgbotapi.NewInlineKeyboardButtonData("Next page", GetPaginationCallbackData(currentPage+1)))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows)
}

func GetPaginationCallbackData(pageNumber int) string {
	return "irregular_verbs_page_" + strconv.Itoa(pageNumber)
}

func GetCurrentPage(chatID int64) int {
	if page, ok := userContext[chatID]; ok {
		return page
	}
	return 1
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
	userContext[chatID] = pageNumber
}
