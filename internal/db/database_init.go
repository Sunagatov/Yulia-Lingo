package db

import (
	"database/sql"
	"fmt"
	"log"
)

var dbConnection *sql.DB

type IrregularVerb struct {
	ID    int    `json:"id"`
	Verb  string `json:"verb"`
	Past  string `json:"past"`
	PastP string `json:"past_participle"`
}

func InitDatabase(db *sql.DB) error {
	dbConnection = db
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS irregular_verbs (
		id SERIAL PRIMARY KEY,
		verb VARCHAR(255),
		past VARCHAR(255),
		past_participle VARCHAR(255)
	)
	`

	_, err := db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("creating database table was failed: %v", err)
	}

	insertQuery := "INSERT INTO irregular_verbs (verb, past, past_participle) VALUES ($1, $2, $3)"

	for _, v := range irregularVerbs {
		split := getVerbParts(v)
		_, err := db.Exec(insertQuery, split[0], split[1], split[2])
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func getVerbParts(verbString string) []string {
	verbParts := make([]string, 3)
	fmt.Sscanf(verbString, "%s - [%s - %s - %s]", &verbParts[0], &verbParts[1], &verbParts[2])
	return verbParts
}
