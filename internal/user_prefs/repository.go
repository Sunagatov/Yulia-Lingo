package user_prefs

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	createTableQuery = `
		CREATE TABLE IF NOT EXISTS user_preferences (
			user_id   BIGINT PRIMARY KEY,
			language  VARCHAR(10) NOT NULL DEFAULT 'ru'
		)`
	upsertLangQuery = `
		INSERT INTO user_preferences (user_id, language) VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET language = $2`
	getLangQuery = `SELECT language FROM user_preferences WHERE user_id = $1`
)

type Repository interface {
	Initialize(ctx context.Context) error
	SetLanguage(ctx context.Context, userID int64, lang string) error
	GetLanguage(ctx context.Context, userID int64) (string, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Initialize(ctx context.Context) error {
	if _, err := r.db.Exec(ctx, createTableQuery); err != nil {
		return fmt.Errorf("create user_preferences table: %w", err)
	}
	return nil
}

func (r *repository) SetLanguage(ctx context.Context, userID int64, lang string) error {
	if _, err := r.db.Exec(ctx, upsertLangQuery, userID, lang); err != nil {
		return fmt.Errorf("upsert language: %w", err)
	}
	return nil
}

func (r *repository) GetLanguage(ctx context.Context, userID int64) (string, error) {
	var lang string
	err := r.db.QueryRow(ctx, getLangQuery, userID).Scan(&lang)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get language: %w", err)
	}
	return lang, nil
}
