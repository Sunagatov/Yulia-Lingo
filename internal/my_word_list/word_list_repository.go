package my_word_list

import (
	"context"
	"fmt"

	"Yulia-Lingo/internal/database"
	"Yulia-Lingo/internal/logger"
)

const (
	getTotalCountQuery = "SELECT COUNT(*) FROM words w JOIN translations t ON w.id = t.word_id WHERE w.part_of_speech = $1"
	getPageQuery       = `
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
	GetTotalCount(ctx context.Context, partOfSpeech string) (int, error)
	GetPage(ctx context.Context, offset, limit int, partOfSpeech string) ([]Entity, error)
	Initialize(ctx context.Context) error
}

type repository struct {
	log logger.Logger
}

func NewRepository(log logger.Logger) Repository {
	return &repository{log: log}
}

func (r *repository) GetTotalCount(ctx context.Context, partOfSpeech string) (int, error) {
	pool, err := database.GetDB()
	if err != nil {
		return 0, fmt.Errorf("failed to get database connection: %w", err)
	}

	var count int
	err = pool.QueryRow(ctx, getTotalCountQuery, partOfSpeech).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return count, nil
}

func (r *repository) GetPage(ctx context.Context, offset, limit int, partOfSpeech string) ([]Entity, error) {
	pool, err := database.GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	rows, err := pool.Query(ctx, getPageQuery, partOfSpeech, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var entities []Entity
	for rows.Next() {
		var entity Entity
		err := rows.Scan(&entity.ID, &entity.Word, &entity.PartOfSpeech, &entity.Translation)
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
	r.log.Info(ctx, "Initializing word list tables")

	pool, err := database.GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if _, err := pool.Exec(ctx, createWordsTableQuery); err != nil {
		return fmt.Errorf("failed to create words table: %w", err)
	}

	if _, err := pool.Exec(ctx, createTranslationsTableQuery); err != nil {
		return fmt.Errorf("failed to create translations table: %w", err)
	}

	r.log.Info(ctx, "Word list tables initialized successfully")
	return nil
}