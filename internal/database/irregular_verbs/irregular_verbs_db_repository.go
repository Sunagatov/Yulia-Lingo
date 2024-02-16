package irregular_verbs

import (
	"Yulia-Lingo/internal/database"
	"fmt"
	"github.com/tealeg/xlsx"
	"log"
	"os"
	"strings"
)

const (
	GetTotalIrregularVerbsCountSqlQuery = "SELECT COUNT(*) FROM irregular_verbs WHERE verb LIKE $1 || '%'"
	GetIrregularVerbsListPageSqlQuery   = "SELECT id, original, verb, past, past_participle FROM irregular_verbs WHERE verb LIKE $3 || '%' LIMIT $1 OFFSET $2"
	DropIrregularVerbsTableSqlQuery     = "DROP TABLE IF EXISTS irregular_verbs"
	InsertSqlQuery                      = "INSERT INTO irregular_verbs (original, verb, past, past_participle) VALUES "
	CreateIrregularVerbsTableSqlQuery   = `
	CREATE TABLE IF NOT EXISTS irregular_verbs (
		id SERIAL PRIMARY KEY,
		Original VARCHAR(255),
		verb VARCHAR(255),
		past VARCHAR(255),
		past_participle VARCHAR(255)
	)
	`
)

type IrregularVerbEntity struct {
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

func GetIrregularVerbsListPage(offset, limit int, letter string) ([]IrregularVerbEntity, error) {
	db, err := database.GetPostgresClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the postgres database: %v", err)
	}

	irregularVerbsDatabaseRows, err := db.Query(GetIrregularVerbsListPageSqlQuery, limit, offset, strings.ToLower(letter))
	if err != nil {
		return nil, fmt.Errorf("failed to execute the database GetIrregularVerbsListPageSqlQuery: %v", err)
	}
	defer irregularVerbsDatabaseRows.Close()

	var irregularVerbsListPage []IrregularVerbEntity

	for irregularVerbsDatabaseRows.Next() {
		var irregularVerb IrregularVerbEntity
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

func InitIrregularVerbsTable() error {
	log.Println("Connecting to database table...")
	db, err := database.GetPostgresClient()
	if err != nil {
		return fmt.Errorf("failed to connect postgres database: %v", err)
	}

	log.Println("Dropping existing table...")
	_, err = db.Exec(DropIrregularVerbsTableSqlQuery)
	if err != nil {
		return fmt.Errorf("failed to drop existing database table: %v", err)
	}
	log.Println("Database table was dropped successfully.")

	log.Println("Creating new irregular verbs database table...")
	_, err = db.Exec(CreateIrregularVerbsTableSqlQuery)
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

		sb.WriteString(InsertSqlQuery)
		sb.WriteString(args)
		sb.WriteString(";\n")
	}

	return sb.String(), nil
}

func readIrregularVerbsXlsxFile() ([]IrregularVerbEntity, error) {
	var irregularVerbs []IrregularVerbEntity

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
		irregularVerb := IrregularVerbEntity{
			Verb:           row.Cells[1].String(),
			Past:           row.Cells[2].String(),
			PastParticiple: row.Cells[3].String(),
			Original:       row.Cells[4].String(),
		}
		irregularVerbs = append(irregularVerbs, irregularVerb)
	}
	return irregularVerbs, nil
}
