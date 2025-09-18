package my_word_list

import (
	"context"
	"fmt"

	"Yulia-Lingo/internal/database"
	"Yulia-Lingo/internal/logger"
)

const (
	getTotalCountQuery = "SELECT COUNT(*) FROM user_words WHERE user_id = $1 AND part_of_speech = $2"
	getPageQuery       = `
		SELECT id, word, part_of_speech, translation 
		FROM user_words 
		WHERE user_id = $1 AND part_of_speech = $2 
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`
	addWordQuery = `
		INSERT INTO user_words (user_id, word, part_of_speech, translation) 
		VALUES ($1, $2, $3, $4) 
		ON CONFLICT (user_id, word, part_of_speech) DO NOTHING
		RETURNING id`
	removeWordQuery           = "DELETE FROM user_words WHERE id = $1 AND user_id = $2"
	createUserWordsTableQuery = `
		CREATE TABLE IF NOT EXISTS user_words (
			id SERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			word VARCHAR(255) NOT NULL,
			part_of_speech VARCHAR(50) NOT NULL,
			translation TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(user_id, word, part_of_speech)
		)`
	createUserWordsIndexQuery = "CREATE INDEX IF NOT EXISTS idx_user_words_user_id_pos ON user_words(user_id, part_of_speech)"
)

type Repository interface {
	GetTotalCount(ctx context.Context, userID int64, partOfSpeech string) (int, error)
	GetPage(ctx context.Context, userID int64, offset, limit int, partOfSpeech string) ([]Entity, error)
	AddWord(ctx context.Context, userID int64, word, partOfSpeech, translation string) error
	RemoveWord(ctx context.Context, userID int64, wordID int) error
	Initialize(ctx context.Context) error
}

type repository struct {
	log logger.Logger
}

func NewRepository(log logger.Logger) Repository {
	return &repository{log: log}
}

func (r *repository) GetTotalCount(ctx context.Context, userID int64, partOfSpeech string) (int, error) {
	pool, err := database.GetDB()
	if err != nil {
		return 0, fmt.Errorf("failed to get database connection: %w", err)
	}

	var count int
	err = pool.QueryRow(ctx, getTotalCountQuery, userID, partOfSpeech).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return count, nil
}

func (r *repository) GetPage(ctx context.Context, userID int64, offset, limit int, partOfSpeech string) ([]Entity, error) {
	pool, err := database.GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	rows, err := pool.Query(ctx, getPageQuery, userID, partOfSpeech, limit, offset)
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
		entity.UserID = userID
		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return entities, nil
}

func (r *repository) AddWord(ctx context.Context, userID int64, word, partOfSpeech, translation string) error {
	pool, err := database.GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	var id int
	err = pool.QueryRow(ctx, addWordQuery, userID, word, partOfSpeech, translation).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to add word: %w", err)
	}

	return nil
}

func (r *repository) RemoveWord(ctx context.Context, userID int64, wordID int) error {
	pool, err := database.GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	_, err = pool.Exec(ctx, removeWordQuery, wordID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove word: %w", err)
	}

	return nil
}

func (r *repository) Initialize(ctx context.Context) error {
	r.log.Info(ctx, "Initializing user word list tables")

	pool, err := database.GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if _, err := pool.Exec(ctx, createUserWordsTableQuery); err != nil {
		return fmt.Errorf("failed to create user_words table: %w", err)
	}

	if _, err := pool.Exec(ctx, createUserWordsIndexQuery); err != nil {
		return fmt.Errorf("failed to create user_words index: %w", err)
	}

	r.log.Info(ctx, "User word list tables initialized successfully")
	return nil
}
