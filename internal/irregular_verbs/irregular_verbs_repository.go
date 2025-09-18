package irregular_verbs

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/database"
	"Yulia-Lingo/internal/logger"

	"github.com/jackc/pgx/v5"
	"github.com/xuri/excelize/v2"
)

const (
	getTotalCountQuery = "SELECT COUNT(*) FROM irregular_verbs WHERE verb LIKE $1 || '%'"
	getPageQuery       = "SELECT id, original, verb, past, past_participle FROM irregular_verbs WHERE verb LIKE $1 || '%' ORDER BY verb LIMIT $2 OFFSET $3"
	dropTableQuery     = "DROP TABLE IF EXISTS irregular_verbs CASCADE"
	createTableQuery   = `
		CREATE TABLE IF NOT EXISTS irregular_verbs (
			id SERIAL PRIMARY KEY,
			original VARCHAR(255) NOT NULL,
			verb VARCHAR(255) NOT NULL UNIQUE,
			past VARCHAR(255) NOT NULL,
			past_participle VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`
	createIndexQuery = "CREATE INDEX IF NOT EXISTS idx_irregular_verbs_verb ON irregular_verbs(verb)"
	insertVerbQuery  = "INSERT INTO irregular_verbs (original, verb, past, past_participle) VALUES ($1, $2, $3, $4) ON CONFLICT (verb) DO NOTHING"
)

type Repository interface {
	GetTotalCount(ctx context.Context, letter string) (int, error)
	GetPage(ctx context.Context, offset, limit int, letter string) ([]Entity, error)
	Initialize(ctx context.Context) error
}

type repository struct {
	cfg *config.Config
	log logger.Logger
}

func NewRepository(cfg *config.Config, log logger.Logger) Repository {
	return &repository{
		cfg: cfg,
		log: log,
	}
}

func (r *repository) GetTotalCount(ctx context.Context, letter string) (int, error) {
	if err := r.validateLetter(letter); err != nil {
		return 0, fmt.Errorf("invalid letter: %w", err)
	}

	db, err := database.GetDB()
	if err != nil {
		return 0, fmt.Errorf("failed to get database connection: %w", err)
	}

	var count int
	err = db.QueryRow(ctx, getTotalCountQuery, strings.ToLower(letter)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return count, nil
}

func (r *repository) GetPage(ctx context.Context, offset, limit int, letter string) ([]Entity, error) {
	if err := r.validateLetter(letter); err != nil {
		return nil, fmt.Errorf("invalid letter: %w", err)
	}

	if offset < 0 || limit <= 0 || limit > 100 {
		return nil, fmt.Errorf("invalid pagination parameters: offset=%d, limit=%d", offset, limit)
	}

	db, err := database.GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	rows, err := db.Query(ctx, getPageQuery, strings.ToLower(letter), limit, offset)
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

func (r *repository) Initialize(ctx context.Context) error {
	r.log.Info(ctx, "Initializing irregular verbs table")

	db, err := database.GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, dropTableQuery); err != nil {
		return fmt.Errorf("failed to drop existing table: %w", err)
	}

	if _, err := tx.Exec(ctx, createTableQuery); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	if _, err := tx.Exec(ctx, createIndexQuery); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	if err := r.insertVerbsFromFile(ctx, tx); err != nil {
		return fmt.Errorf("failed to insert verbs: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.log.Info(ctx, "Irregular verbs table initialized successfully")
	return nil
}

func (r *repository) insertVerbsFromFile(ctx context.Context, tx pgx.Tx) error {
	entities, err := r.readFromFile(ctx)
	if err != nil {
		return fmt.Errorf("failed to read from file: %w", err)
	}

	if len(entities) == 0 {
		return fmt.Errorf("no irregular verbs found in file")
	}

	for _, entity := range entities {
		entity.Sanitize()
		if err := entity.Validate(); err != nil {
			r.log.Warn(ctx, "Skipping invalid entity",
				logger.Field{Key: "verb", Value: entity.Verb},
				logger.Field{Key: "error", Value: err.Error()},
			)
			continue
		}

		_, err := tx.Exec(ctx, insertVerbQuery,
			entity.Original, entity.Verb, entity.Past, entity.PastParticiple)
		if err != nil {
			return fmt.Errorf("failed to insert verb %s: %w", entity.Verb, err)
		}
	}

	return nil
}

func (r *repository) readFromFile(ctx context.Context) ([]Entity, error) {
	filePath, err := r.cfg.GetIrregularVerbsFilePath(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get file path: %w", err)
	}

	file, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer file.Close()

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	rows, err := file.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	var entities []Entity
	for i, row := range rows {
		if i == 0 || len(row) < 5 { // Skip header and incomplete rows
			continue
		}

		entity := Entity{
			Verb:           row[1],
			Past:           row[2],
			PastParticiple: row[3],
			Original:       row[4],
		}

		if strings.TrimSpace(entity.Verb) == "" {
			break
		}

		entities = append(entities, entity)
	}

	r.log.Info(ctx, "Loaded irregular verbs from file",
		logger.Field{Key: "count", Value: len(entities)},
		logger.Field{Key: "file_path", Value: filePath},
	)
	return entities, nil
}

func (r *repository) validateLetter(letter string) error {
	if len(letter) != 1 {
		return fmt.Errorf("letter must be exactly one character")
	}
	if !unicode.IsLetter(rune(letter[0])) {
		return fmt.Errorf("letter must be alphabetic")
	}
	return nil
}