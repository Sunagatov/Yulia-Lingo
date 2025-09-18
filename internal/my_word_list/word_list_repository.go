package my_word_list

import (
	"fmt"

	"Yulia-Lingo/internal/database"
)

const (
	getTotalCountQuery = "SELECT COUNT(*) FROM words w JOIN translations t ON w.id = t.word_id WHERE w.part_of_speech = $1"
	getPageQuery       = `
		SELECT w.id, w.word, w.part_of_speech, t.translation 
		FROM words w 
		JOIN translations t ON w.id = t.word_id 
		WHERE w.part_of_speech = $3 
		LIMIT $1 OFFSET $2`
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
	GetTotalCount(partOfSpeech string) (int, error)
	GetPage(offset, limit int, partOfSpeech string) ([]Entity, error)
	Initialize() error
}

type repository struct{}

func NewRepository() Repository {
	return &repository{}
}

func (r *repository) GetTotalCount(partOfSpeech string) (int, error) {
	db, err := database.GetDB()
	if err != nil {
		return 0, fmt.Errorf("failed to get database connection: %w", err)
	}

	var count int
	err = db.QueryRow(getTotalCountQuery, partOfSpeech).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return count, nil
}

func (r *repository) GetPage(offset, limit int, partOfSpeech string) ([]Entity, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	rows, err := db.Query(getPageQuery, limit, offset, partOfSpeech)
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

func (r *repository) Initialize() error {
	// Initializing word list tables

	db, err := database.GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if _, err := db.Exec(createWordsTableQuery); err != nil {
		return fmt.Errorf("failed to create words table: %w", err)
	}

	if _, err := db.Exec(createTranslationsTableQuery); err != nil {
		return fmt.Errorf("failed to create translations table: %w", err)
	}

	// Word list tables initialized successfully
	return nil
}