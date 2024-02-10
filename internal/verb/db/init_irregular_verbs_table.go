package db

import (
	db2 "Yulia-Lingo/internal/db"
	"fmt"
	"log"
)

func InitIrregularVerbsTable() error {
	db, err := db2.GetPostgresClient()
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
		Original VARCHAR(255),
		verb VARCHAR(255),
		past VARCHAR(255),
		past_participle VARCHAR(255)
	)
	`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("creating database table failed: %v", err)
	}
	log.Println("Table created successfully.")

	runSaveDataFromFiles()
	return nil
}

func runSaveDataFromFiles() {
	err := SaveIrregularVerbs()
	if err != nil {
		log.Printf("Can't insert irregular verbs data, err: %v", err)
	}
}
