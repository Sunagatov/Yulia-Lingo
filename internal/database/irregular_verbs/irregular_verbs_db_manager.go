package irregular_verbs

import (
	"Yulia-Lingo/internal/database"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

const (
	IrregularVerbsCountPerPage          = 7
	GetTotalIrregularVerbsCountSqlQuery = "SELECT COUNT(*) FROM irregular_verbs WHERE verb LIKE $1 || '%'"
	GetIrregularVerbsListPageSqlQuery   = "SELECT id, original, verb, past, past_participle FROM irregular_verbs WHERE verb LIKE $3 || '%' LIMIT $1 OFFSET $2"
)

type KeyboardVerbValue struct {
	Request string
	Page    int
	Latter  string
}

type IrregularVerb struct {
	ID             int    `json:"id"`
	Original       string `json:"original"`
	Verb           string `json:"verb"`
	Past           string `json:"past"`
	PastParticiple string `json:"past_participle"`
}

func GetTotalIrregularVerbsCount(letter string) (int, error) {
	db, err := database.GetPostgresClient()
	if err != nil {
		return -1, fmt.Errorf("failed to connect to the postgres database: %v", err)
	}

	preparedSqlStatement, err := db.Prepare(GetTotalIrregularVerbsCountSqlQuery)
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
	offset := currentPageNumber * IrregularVerbsCountPerPage
	irregularVerbsListPage, err := GetIrregularVerbsListPage(offset, IrregularVerbsCountPerPage, selectedLetter)
	if err != nil {
		return "", fmt.Errorf("failed to get irregularVerbs page from database: %v", err)
	}
	if len(irregularVerbsListPage) == 0 {
		return "", nil
	}
	var irregularVerbsPageAsText string
	for _, verb := range irregularVerbsListPage {
		irregularVerbsPageAsText += fmt.Sprintf("*%s* - *[*%s - %s - %s*]*\n\n", verb.Original, verb.Verb, verb.Past, verb.PastParticiple)
	}
	return irregularVerbsPageAsText, nil
}

func GetIrregularVerbsListPage(offset, limit int, letter string) ([]IrregularVerb, error) {
	db, err := database.GetPostgresClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the postgres database: %v", err)
	}

	irregularVerbsDatabaseRows, err := db.Query(GetIrregularVerbsListPageSqlQuery, limit, offset, strings.ToLower(letter))
	if err != nil {
		return nil, fmt.Errorf("failed to execute the database GetIrregularVerbsListPageSqlQuery: %v", err)
	}
	defer irregularVerbsDatabaseRows.Close()

	var irregularVerbsListPage []IrregularVerb

	for irregularVerbsDatabaseRows.Next() {
		var irregularVerb IrregularVerb
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

func ConvertToJson(entity interface{}) (string, error) {
	jsonBytes, err := json.Marshal(entity)
	if err != nil {
		return "", fmt.Errorf("failed to create a json: %v", err)
	}
	return string(jsonBytes), nil
}

func KeyboardVerbValueFromJSON(jsonStr string) (KeyboardVerbValue, error) {
	var kv KeyboardVerbValue
	err := json.Unmarshal([]byte(jsonStr), &kv)
	if err != nil {
		return KeyboardVerbValue{}, err
	}
	return kv, nil
}

func CreateInlineKeyboard(currentPage int, letter string) ([]tgbotapi.InlineKeyboardButton, error) {
	totalVerbs, err := GetTotalIrregularVerbsCount(letter)
	if err != nil {
		return nil, fmt.Errorf("failed to get total irregular verbs count: %v", err)
	}
	totalPages := totalVerbs / IrregularVerbsCountPerPage

	var keyboard []tgbotapi.InlineKeyboardButton
	if currentPage > 0 {
		jsonPrev, err := ConvertToJson(KeyboardVerbValue{
			Request: "IrregularVerbs",
			Page:    currentPage - 1,
			Latter:  letter,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create a json for the case (currentPage > 0): %v", err)
		}
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardButtonData("Prev page", jsonPrev))
	}
	if currentPage < totalPages && totalVerbs > IrregularVerbsCountPerPage {
		jsonNext, err := ConvertToJson(KeyboardVerbValue{
			Request: "IrregularVerbs",
			Page:    currentPage + 1,
			Latter:  letter,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create a json for the case (currentPage < totalPages): %v", err)
		}
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardButtonData("Next page", jsonNext))
	}

	if len(keyboard) == 0 {
		return nil, nil
	}

	return keyboard, nil
}
