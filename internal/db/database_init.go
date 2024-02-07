package db

import (
	"fmt"
	"log"
	"strings"
)

// IrregularVerb represents the structure of irregular verbs
type IrregularVerb struct {
	ID             int    `json:"id"`
	Translated     string `json:"translated"`
	Original       string `json:"original"`
	Past           string `json:"past"`
	PastParticiple string `json:"past_participle"`
}

// InitDatabase initializes the database and inserts irregular verbs data
func InitDatabase() error {
	db, err := GetPostgresClient()
	if err != nil {
		return fmt.Errorf("database wosn't conecteedm, err: %v", err)
	}

	log.Println("Dropping existing table...")
	_, err = db.Exec("DROP TABLE IF EXISTS irregular_verbs")
	if err != nil {
		return fmt.Errorf("dropping existing table failed: %v", err)
	}
	log.Println("Table dropped successfully.")

	log.Println("Creating new table...")
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS irregular_verbs (
		id SERIAL PRIMARY KEY,
		translated VARCHAR(255),
		original VARCHAR(255),
		past VARCHAR(255),
		past_participle VARCHAR(255)
	)
	`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("creating database table failed: %v", err)
	}
	log.Println("Table created successfully.")

	log.Println("Inserting irregular verbs data...")
	insertQuery := "INSERT INTO irregular_verbs (translated, original, past, past_participle) VALUES ($1, $2, $3, $4)"

	for _, irregularVerb := range irregularVerbs {
		irregularVerbParts, err := getIrregularVerbParts(irregularVerb)
		if err != nil {
			return fmt.Errorf("getting verb parts from database table failed: %v", err)
		}

		_, err = db.Exec(insertQuery, irregularVerbParts.Translated, irregularVerbParts.Original, irregularVerbParts.Past, irregularVerbParts.PastParticiple)
		if err != nil {
			return fmt.Errorf("database initialization failed: %v", err)
		}
	}
	log.Println("Irregular verbs data inserted successfully.")

	log.Println("Database initialization completed successfully.")
	return nil
}

// getIrregularVerbParts parses the irregular verb string and returns its parts
func getIrregularVerbParts(irregularVerb string) (*IrregularVerb, error) {
	parts := strings.Split(irregularVerb, ";")
	if len(parts) != 2 {
		return nil, fmt.Errorf("failed to parse irregular verb format")
	}

	original := parts[0]
	verbParts := strings.Split(parts[1], " - ")

	if len(verbParts) != 3 {
		return nil, fmt.Errorf("failed to parse irregular verb format")
	}

	return &IrregularVerb{
		Original:       original,
		Past:           verbParts[0],
		PastParticiple: verbParts[1],
		Translated:     verbParts[2],
	}, nil
}
