package database

import (
	"context"
	"fmt"

	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, cfg *config.Config, log logger.Logger) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password,
		cfg.Database.Host, cfg.Database.Port,
		cfg.Database.Name, cfg.Database.SSLMode,
	)
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}
	poolConfig.MaxConns = int32(cfg.Database.MaxConns)
	poolConfig.MinConns = int32(cfg.Database.MinConns)
	poolConfig.MaxConnLifetime = cfg.Database.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.Database.MaxConnIdleTime
	poolConfig.ConnConfig.ConnectTimeout = cfg.Database.ConnectTimeout
	poolConfig.ConnConfig.RuntimeParams = map[string]string{"application_name": cfg.Database.ApplicationName}

	ctx, cancel := context.WithTimeout(ctx, cfg.Database.PingTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	log.Info(ctx, "db.connected",
		logger.Field{Key: "max_conns", Value: cfg.Database.MaxConns},
		logger.Field{Key: "min_conns", Value: cfg.Database.MinConns},
	)
	return pool, nil
}
