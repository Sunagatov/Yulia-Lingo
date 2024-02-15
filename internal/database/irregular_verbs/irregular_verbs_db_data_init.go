package irregular_verbs

import (
	"Yulia-Lingo/internal/database"
	"Yulia-Lingo/internal/database/irregular_verbs/model"
	"fmt"
	"github.com/tealeg/xlsx"
	"log"
	"os"
	"strings"
)

const (
	dropIrregularVerbsTableSqlQuery   = "DROP TABLE IF EXISTS irregular_verbs"
	insertSqlQuery                    = "INSERT INTO irregular_verbs (original, verb, past, past_participle) VALUES "
	createIrregularVerbsTableSqlQuery = `
	CREATE TABLE IF NOT EXISTS irregular_verbs (
		id SERIAL PRIMARY KEY,
		Original VARCHAR(255),
		verb VARCHAR(255),
		past VARCHAR(255),
		past_participle VARCHAR(255)
	)
	`
)

func InitIrregularVerbsTable() error {
	log.Println("Connecting to database table...")
	db, err := database.GetPostgresClient()
	if err != nil {
		return fmt.Errorf("failed to connect postgres database: %v", err)
	}

	log.Println("Dropping existing table...")
	_, err = db.Exec(dropIrregularVerbsTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to drop existing database table: %v", err)
	}
	log.Println("Database table was dropped successfully.")

	log.Println("Creating new irregular verbs database table...")
	_, err = db.Exec(createIrregularVerbsTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to create new irregular verbs database table: %v", err)
	}
	log.Println("Database table for irregular verbs was created successfully.")

	query, err := prepareIrregularVerbsSqlQueryInserts()
	if err != nil {
		return fmt.Errorf("failed to prepare sqlQuery inserts ti fill the irregular verbs database table: %v", err)
	}
	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to insert irregular verbs data to the database table: %v", err)
	}
	return nil
}

func prepareIrregularVerbsSqlQueryInserts() (string, error) {
	irregularVerbsFromFile, err := readIrregularVerbsXlsxFile()
	if err != nil {
		return "", fmt.Errorf("failed to get irregular verbs from the excel file: %v", err)
	}

	var sb strings.Builder
	for _, irregularVerb := range irregularVerbsFromFile {
		args := fmt.Sprintf("(%s, %s, %s, %s)",
			"'"+irregularVerb.Original+"'",
			"'"+irregularVerb.Verb+"'",
			"'"+irregularVerb.Past+"'",
			"'"+irregularVerb.PastParticiple+"'")

		sb.WriteString(insertSqlQuery)
		sb.WriteString(args)
		sb.WriteString(";\n")
	}

	return sb.String(), nil
}

func readIrregularVerbsXlsxFile() ([]model.IrregularVerb, error) {
	var irregularVerbs []model.IrregularVerb

	irregularVerbsFilePath := os.Getenv("IRREGULAR_VERBS_FILE_PATH")
	if irregularVerbsFilePath == "" {
		log.Fatalf("failed to read IRREGULAR_VERBS_FILE_PATH from the environment variables")
	}

	excelFile, err := xlsx.OpenFile(irregularVerbsFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open xlsx file with irregular verbs dataset: %v", err)
	}

	irregularVerbsExcelTable := excelFile.Sheets[0]

	for i, row := range irregularVerbsExcelTable.Rows {
		if i == 0 {
			continue
		}
		if len(row.Cells) == 0 {
			break
		}
		irregularVerb := model.IrregularVerb{
			Verb:           row.Cells[1].String(),
			Past:           row.Cells[2].String(),
			PastParticiple: row.Cells[3].String(),
			Original:       row.Cells[4].String(),
		}
		irregularVerbs = append(irregularVerbs, irregularVerb)
	}
	return irregularVerbs, nil
}
