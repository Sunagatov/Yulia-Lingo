package database

import (
	"context"
	"fmt"
	"time"

	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

func Initialize(ctx context.Context, cfg *config.Config, log logger.Logger) error {
	poolConfig, err := pgxpool.ParseConfig(cfg.GetDatabaseURL())
	if err != nil {
		return fmt.Errorf("parse db config: %w", err)
	}
	poolConfig.MaxConns = int32(cfg.Database.MaxConns)
	poolConfig.MinConns = int32(cfg.Database.MinConns)
	poolConfig.MaxConnLifetime = cfg.Database.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.Database.MaxConnIdleTime
	poolConfig.ConnConfig.ConnectTimeout = 10 * time.Second
	poolConfig.ConnConfig.RuntimeParams = map[string]string{"application_name": "yulia-lingo"}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	p, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}
	if err := p.Ping(ctx); err != nil {
		p.Close()
		return fmt.Errorf("ping db: %w", err)
	}
	pool = p
	log.Info(ctx, "db.pool_established",
		logger.Field{Key: "max_conns", Value: cfg.Database.MaxConns},
		logger.Field{Key: "min_conns", Value: cfg.Database.MinConns},
	)
	return nil
}

func GetDB() (*pgxpool.Pool, error) {
	if pool == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return pool, nil
}

func Close() {
	if pool != nil {
		pool.Close()
	}
}

func HealthCheck(ctx context.Context) error {
	p, err := GetDB()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return p.Ping(ctx)
}
