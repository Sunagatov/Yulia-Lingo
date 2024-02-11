package repository

import (
	database "Yulia-Lingo/internal/db"
	"Yulia-Lingo/internal/verb/model"
	"database/sql"
	"fmt"
	"strings"
)

func GetVerbsListFromLatter(letter string, page, pageSize int64) ([]model.IrregularVerb, error) {
	db, err := database.GetPostgresClient()
	if err != nil {
		return nil, fmt.Errorf("can't connect to postgres, err: %v", err)
	}

	skipSize := page * pageSize
	query := `
			SELECT * FROM irregular_verbs
      		WHERE first_letter = $1
        	LIMIT $2 OFFSET $3
		`

	rows, err := db.Query(query, strings.ToLower(letter), pageSize, skipSize)
	if err != nil {
		return nil, fmt.Errorf("error executing database query: %v", err)
	}

	return convertRowToIrregularVerbs(rows)
}

func convertRowToIrregularVerbs(rows *sql.Rows) ([]model.IrregularVerb, error) {
	var verbs []model.IrregularVerb

	for rows.Next() {
		var verb model.IrregularVerb
		err := rows.Scan(&verb.ID, &verb.Original, &verb.Verb, &verb.FirstLetter, &verb.Past, &verb.PastP)
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

func GetTotalIrregularVerbsCount(letter string) (int, error) {
	db, err := database.GetPostgresClient()
	if err != nil {
		return 0, fmt.Errorf("can't connect to postgres, err: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM irregular_verbs WHERE first_letter = $1", strings.ToLower(letter)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting total irregular verbs count: %v", err)
	}
	return count, err
}
