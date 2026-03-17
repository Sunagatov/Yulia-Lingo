package openai

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UsageRepository struct {
	pool *pgxpool.Pool
}

func NewUsageRepository(pool *pgxpool.Pool) *UsageRepository {
	return &UsageRepository{pool: pool}
}

func (r *UsageRepository) GetUserDailyCount(userID int64) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM ai_usage_log 
		WHERE user_id = $1 AND timestamp > NOW() - INTERVAL '24 hours'
	`
	var count int
	err := r.pool.QueryRow(context.Background(), query, userID).Scan(&count)
	return count, err
}

func (r *UsageRepository) GetUserHourlyCount(userID int64) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM ai_usage_log 
		WHERE user_id = $1 AND timestamp > NOW() - INTERVAL '1 hour'
	`
	var count int
	err := r.pool.QueryRow(context.Background(), query, userID).Scan(&count)
	return count, err
}

func (r *UsageRepository) GetGlobalDailyCount() (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM ai_usage_log 
		WHERE timestamp > NOW() - INTERVAL '24 hours'
	`
	var count int
	err := r.pool.QueryRow(context.Background(), query).Scan(&count)
	return count, err
}

func (r *UsageRepository) RecordUsage(userID int64, feature string) error {
	query := `
		INSERT INTO ai_usage_log (user_id, feature, timestamp)
		VALUES ($1, $2, $3)
	`
	_, err := r.pool.Exec(context.Background(), query, userID, feature, time.Now())
	return err
}
