package irregular_verbs

import (
	"fmt"
	"strings"

	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/database"
	"Yulia-Lingo/internal/logger"

	"github.com/sirupsen/logrus"
	"github.com/tealeg/xlsx"
)

const (
	getTotalCountQuery = "SELECT COUNT(*) FROM irregular_verbs WHERE verb LIKE $1 || '%'"
	getPageQuery       = "SELECT id, original, verb, past, past_participle FROM irregular_verbs WHERE verb LIKE $3 || '%' LIMIT $1 OFFSET $2"
	dropTableQuery     = "DROP TABLE IF EXISTS irregular_verbs"
	createTableQuery   = `
		CREATE TABLE IF NOT EXISTS irregular_verbs (
			id SERIAL PRIMARY KEY,
			original VARCHAR(255),
			verb VARCHAR(255),
			past VARCHAR(255),
			past_participle VARCHAR(255)
		)`
)

type Repository interface {
	GetTotalCount(letter string) (int, error)
	GetPage(offset, limit int, letter string) ([]Entity, error)
	Initialize() error
}

type repository struct {
	cfg *config.Config
}

func NewRepository(cfg *config.Config) Repository {
	return &repository{cfg: cfg}
}

func (r *repository) GetTotalCount(letter string) (int, error) {
	db, err := database.GetDB()
	if err != nil {
		return 0, fmt.Errorf("failed to get database connection: %w", err)
	}

	stmt, err := db.Prepare(getTotalCountQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var count int
	err = stmt.QueryRow(strings.ToLower(letter)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return count, nil
}

func (r *repository) GetPage(offset, limit int, letter string) ([]Entity, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	rows, err := db.Query(getPageQuery, limit, offset, strings.ToLower(letter))
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var entities []Entity
	for rows.Next() {
		var entity Entity
		err := rows.Scan(&entity.ID, &entity.Original, &entity.Verb, &entity.Past, &entity.PastParticiple)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return entities, nil
}

func (r *repository) Initialize() error {
	logger.Info("Initializing irregular verbs table")

	db, err := database.GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if _, err := db.Exec(dropTableQuery); err != nil {
		return fmt.Errorf("failed to drop existing table: %w", err)
	}

	if _, err := db.Exec(createTableQuery); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	insertQuery, err := r.prepareInsertQuery()
	if err != nil {
		return fmt.Errorf("failed to prepare insert query: %w", err)
	}

	if _, err := db.Exec(insertQuery); err != nil {
		return fmt.Errorf("failed to insert data: %w", err)
	}

	logger.Info("Irregular verbs table initialized successfully")
	return nil
}

func (r *repository) prepareInsertQuery() (string, error) {
	entities, err := r.readFromFile()
	if err != nil {
		return "", fmt.Errorf("failed to read from file: %w", err)
	}

	if len(entities) == 0 {
		return "", fmt.Errorf("no irregular verbs found in file")
	}

	var values []string
	for _, entity := range entities {
		value := fmt.Sprintf("('%s', '%s', '%s', '%s')",
			strings.ReplaceAll(entity.Original, "'", "''"),
			strings.ReplaceAll(entity.Verb, "'", "''"),
			strings.ReplaceAll(entity.Past, "'", "''"),
			strings.ReplaceAll(entity.PastParticiple, "'", "''"))
		values = append(values, value)
	}

	return fmt.Sprintf("INSERT INTO irregular_verbs (original, verb, past, past_participle) VALUES %s",
		strings.Join(values, ", ")), nil
}

func (r *repository) readFromFile() ([]Entity, error) {
	filePath, err := r.cfg.GetIrregularVerbsFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get file path: %w", err)
	}

	file, err := xlsx.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}

	if len(file.Sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	sheet := file.Sheets[0]
	var entities []Entity

	for i, row := range sheet.Rows {
		if i == 0 || len(row.Cells) < 5 {
			continue
		}

		entity := Entity{
			Verb:           row.Cells[1].String(),
			Past:           row.Cells[2].String(),
			PastParticiple: row.Cells[3].String(),
			Original:       row.Cells[4].String(),
		}

		if entity.Verb == "" {
			break
		}

		entities = append(entities, entity)
	}

	logger.Info("Loaded irregular verbs from file", logrus.Fields{"count": len(entities)})
	return entities, nil
}