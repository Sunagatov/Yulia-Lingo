package my_word_list

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	getCountsQuery = "SELECT part_of_speech, COUNT(*) FROM words GROUP BY part_of_speech"
	getPageQuery   = `
		SELECT w.id, w.word, w.part_of_speech, t.translation 
		FROM words w 
		JOIN translations t ON w.id = t.word_id 
		WHERE w.part_of_speech = $1 
		LIMIT $2 OFFSET $3`
	createWordsTableQuery = `
		CREATE TABLE IF NOT EXISTS words (
			id SERIAL PRIMARY KEY,
			word VARCHAR(255) NOT NULL,
			part_of_speech VARCHAR(50) NOT NULL
		)`
	createTranslationsTableQuery = `
		CREATE TABLE IF NOT EXISTS translations (
			id SERIAL PRIMARY KEY,
			word_id INTEGER REFERENCES words(id),
			translation VARCHAR(255) NOT NULL
		)`
)

type Repository interface {
	GetCountsByPOS(ctx context.Context) (map[string]int, error)
	GetPage(ctx context.Context, offset, limit int, partOfSpeech string) ([]Entity, error)
	Initialize(ctx context.Context) error
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) GetCountsByPOS(ctx context.Context) (map[string]int, error) {
	rows, err := r.db.Query(ctx, getCountsQuery)
	if err != nil {
		return nil, fmt.Errorf("query counts: %w", err)
	}
	defer rows.Close()
	counts := make(map[string]int)
	for rows.Next() {
		var pos string
		var count int
		if err := rows.Scan(&pos, &count); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		counts[pos] = count
	}
	return counts, rows.Err()
}

func (r *repository) GetPage(ctx context.Context, offset, limit int, partOfSpeech string) ([]Entity, error) {
	rows, err := r.db.Query(ctx, getPageQuery, partOfSpeech, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query page: %w", err)
	}
	defer rows.Close()
	var entities []Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Word, &e.PartOfSpeech, &e.Translation); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		entities = append(entities, e)
	}
	return entities, rows.Err()
}

func (r *repository) Initialize(ctx context.Context) error {
	for _, q := range []string{createWordsTableQuery, createTranslationsTableQuery} {
		if _, err := r.db.Exec(ctx, q); err != nil {
			return fmt.Errorf("exec schema: %w", err)
		}
	}
	return nil
}
