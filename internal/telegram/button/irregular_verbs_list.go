package button

import (
	database "Yulia-Lingo/internal/database"
	"Yulia-Lingo/internal/verb/model"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"math"
	"strconv"
	"strings"
)

const (
	IrregularVerbsCountPerPage = 10
)

var userContext = make(map[int64]int)

func GetTotalIrregularVerbsCount(letter string) (int, error) {
	db, err := database.GetPostgresClient()
	if err != nil {
		return -1, fmt.Errorf("failed to connect to the postgres database: %v", err)
	}

	sqlQuery := "SELECT COUNT(*) FROM irregular_verbs WHERE verb LIKE $1 || '%'"
	preparedSqlStatement, err := db.Prepare(sqlQuery)
	if err != nil {
		return -1, fmt.Errorf("failed to prepare sql statement: %v", err)
	}

	var totalIrregularVerbsCount int
	err = preparedSqlStatement.QueryRow(strings.ToLower(letter)).Scan(&totalIrregularVerbsCount)
	if err != nil {
		return -1, fmt.Errorf("failed to execute the sqlQuery for getting totalIrregularVerbsCount from database: %v", err)
	}
	defer preparedSqlStatement.Close()
	return totalIrregularVerbsCount, nil
}

func GetIrregularVerbsPageAsText(currentPageNumber int, selectedLetter string) (string, error) {
	offset := (currentPageNumber - 1) * IrregularVerbsCountPerPage
	irregularVerbsListPage, err := GetIrregularVerbsListPage(offset, IrregularVerbsCountPerPage, selectedLetter)
	if err != nil {
		return "", fmt.Errorf("failed to get irregularVerbs page from database: %v", err)
	}
	var irregularVerbsPageAsText string
	for _, verb := range irregularVerbsListPage {
		irregularVerbsPageAsText += fmt.Sprintf("%s - [%s - %s - %s]\n", verb.Original, verb.Verb, verb.Past, verb.PastParticiple)
	}
	return irregularVerbsPageAsText, nil
}

func GetIrregularVerbsListPage(offset, limit int, letter string) ([]model.IrregularVerb, error) {
	db, err := database.GetPostgresClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the postgres database: %v", err)
	}

	query := "SELECT id, original, verb, past, past_participle FROM irregular_verbs WHERE verb LIKE $3 || '%' LIMIT $1 OFFSET $2"
	irregularVerbsDatabaseRows, err := db.Query(query, limit, offset, strings.ToLower(letter))
	if err != nil {
		return nil, fmt.Errorf("failed to execute the database query: %v", err)
	}
	defer irregularVerbsDatabaseRows.Close()

	var irregularVerbsListPage []model.IrregularVerb

	for irregularVerbsDatabaseRows.Next() {
		var irregularVerb model.IrregularVerb
		err = irregularVerbsDatabaseRows.Scan(&irregularVerb.ID, &irregularVerb.Original, &irregularVerb.Verb, &irregularVerb.Past, &irregularVerb.PastParticiple)
		if err != nil {
			return nil, fmt.Errorf("failed to scan the row: %v", err)
		}
		irregularVerbsListPage = append(irregularVerbsListPage, irregularVerb)
	}
	if irregularVerbsDatabaseRows.Err() != nil {
		return nil, fmt.Errorf("failed to iterate over irregularVerbsDatabaseRows: %v", err)
	}
	return irregularVerbsListPage, nil
}

func CreateInlinePaginationButtonsForIrregularVerbsPage(currentPage int, totalVerbs int, selectedLetter string) tgbotapi.InlineKeyboardMarkup {
	totalPages := int(math.Ceil(float64(totalVerbs) / IrregularVerbsCountPerPage))

	var keyboard []tgbotapi.InlineKeyboardButton
	if currentPage > 1 {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardButtonData("⬅️Назад", GetPaginationCallbackData(currentPage-1, selectedLetter)))
	}
	if currentPage < totalPages && totalVerbs > IrregularVerbsCountPerPage {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardButtonData("Вперед ➡️", GetPaginationCallbackData(currentPage+1, selectedLetter)))
	}
	return tgbotapi.NewInlineKeyboardMarkup(keyboard)
}

func GetPaginationCallbackData(pageNumber int, selectedLetter string) string {
	return "irregular_verbs_page_" + strconv.Itoa(pageNumber) + "_" + selectedLetter
}

func GetCurrentPageNumber(chatID int64) (int, error) {
	pageNumber, ok := userContext[chatID]
	if ok {
		return pageNumber, nil
	} else {
		return -1, fmt.Errorf("failed to retrieve current irregular verbs page for a user")
	}
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
