package irregular_verbs

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"Yulia-Lingo/internal/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"
)

const (
	getTotalCountQuery = "SELECT COUNT(*) FROM irregular_verbs WHERE verb LIKE $1 || '%'"
	getPageQuery       = "SELECT id, original, verb, past, past_participle FROM irregular_verbs WHERE verb LIKE $1 || '%' ORDER BY verb LIMIT $2 OFFSET $3"
	createTableQuery   = `
		CREATE TABLE IF NOT EXISTS irregular_verbs (
			id SERIAL PRIMARY KEY,
			original VARCHAR(255) NOT NULL,
			verb VARCHAR(255) NOT NULL UNIQUE,
			past VARCHAR(255) NOT NULL,
			past_participle VARCHAR(255) NOT NULL
		)`
	createIndexQuery = "CREATE INDEX IF NOT EXISTS idx_irregular_verbs_verb ON irregular_verbs(verb)"
	insertVerbQuery  = "INSERT INTO irregular_verbs (original, verb, past, past_participle) VALUES ($1, $2, $3, $4) ON CONFLICT (verb) DO NOTHING"
)

type Repository interface {
	GetTotalCount(ctx context.Context, letter string) (int, error)
	GetPage(ctx context.Context, offset, limit int, letter string) ([]Entity, error)
	GetLetterCounts(ctx context.Context) (map[string]int, error)
	Initialize(ctx context.Context) error
}

type repository struct {
	db       *pgxpool.Pool
	filePath string
	log      logger.Logger
}

func NewRepository(db *pgxpool.Pool, verbsFilePath string, log logger.Logger) Repository {
	if !filepath.IsAbs(verbsFilePath) {
		if abs, err := filepath.Abs(verbsFilePath); err == nil {
			verbsFilePath = abs
		}
	}
	return &repository{db: db, filePath: verbsFilePath, log: log}
}

func (r *repository) GetTotalCount(ctx context.Context, letter string) (int, error) {
	if !isValidLetter(letter) {
		return 0, fmt.Errorf("invalid letter: %q", letter)
	}
	var count int
	return count, r.db.QueryRow(ctx, getTotalCountQuery, strings.ToLower(letter)).Scan(&count)
}

func (r *repository) GetPage(ctx context.Context, offset, limit int, letter string) ([]Entity, error) {
	if !isValidLetter(letter) {
		return nil, fmt.Errorf("invalid letter: %q", letter)
	}
	rows, err := r.db.Query(ctx, getPageQuery, strings.ToLower(letter), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var entities []Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Original, &e.Verb, &e.Past, &e.PastParticiple); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		entities = append(entities, e)
	}
	return entities, rows.Err()
}

func (r *repository) GetLetterCounts(ctx context.Context) (map[string]int, error) {
	rows, err := r.db.Query(ctx, `SELECT UPPER(LEFT(verb,1)), COUNT(*) FROM irregular_verbs GROUP BY 1 ORDER BY 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := make(map[string]int)
	for rows.Next() {
		var l string
		var c int
		if err := rows.Scan(&l, &c); err == nil {
			counts[l] = c
		}
	}
	return counts, rows.Err()
}

func (r *repository) Initialize(ctx context.Context) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, q := range []string{createTableQuery, createIndexQuery} {
		if _, err := tx.Exec(ctx, q); err != nil {
			return fmt.Errorf("exec schema: %w", err)
		}
	}

	var count int
	if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM irregular_verbs").Scan(&count); err != nil {
		return fmt.Errorf("count verbs: %w", err)
	}
	if count == 0 {
		if err := r.insertVerbsFromFile(ctx, tx); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *repository) insertVerbsFromFile(ctx context.Context, tx pgx.Tx) error {
	entities, err := r.readFromFile(ctx)
	if err != nil {
		return err
	}
	if len(entities) == 0 {
		return fmt.Errorf("no irregular verbs found in file")
	}
	for _, e := range entities {
		e.Original = strings.TrimSpace(e.Original)
		e.Verb = normalise(e.Verb)
		e.Past = normalise(e.Past)
		e.PastParticiple = normalise(e.PastParticiple)
		if e.Verb == "" || e.Past == "" || e.PastParticiple == "" {
			r.log.Warn(ctx, "verb.skip_invalid", logger.Field{Key: "verb", Value: e.Verb})
			continue
		}
		if _, err := tx.Exec(ctx, insertVerbQuery, e.Original, e.Verb, e.Past, e.PastParticiple); err != nil {
			return fmt.Errorf("insert %q: %w", e.Verb, err)
		}
	}
	return nil
}

func normalise(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func (r *repository) readFromFile(ctx context.Context) ([]Entity, error) {
	file, err := excelize.OpenFile(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("open excel: %w", err)
	}
	defer file.Close()

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets in excel file")
	}
	rows, err := file.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("get rows: %w", err)
	}

	var entities []Entity
	for i, row := range rows {
		if i == 0 || len(row) < 5 {
			continue
		}
		if strings.TrimSpace(row[1]) == "" {
			break
		}
		entities = append(entities, Entity{Verb: row[1], Past: row[2], PastParticiple: row[3], Original: row[4]})
	}
	r.log.Info(ctx, "verbs.loaded", logger.Field{Key: "count", Value: len(entities)})
	return entities, nil
}

func isValidLetter(letter string) bool {
	return len(letter) == 1 && unicode.IsLetter(rune(letter[0]))
}
