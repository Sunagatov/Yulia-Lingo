package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (interface{}, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	Ping(ctx context.Context) error
	Close()
}

type Manager struct {
	pool *pgxpool.Pool
	mu   sync.RWMutex
	cfg  *config.Config
	log  logger.Logger
}

var instance *Manager
var once sync.Once

func Initialize(ctx context.Context, cfg *config.Config, log logger.Logger) error {
	var err error
	once.Do(func() {
		instance = &Manager{
			cfg: cfg,
			log: log,
		}
		err = instance.connect(ctx)
	})
	return err
}

func GetDB() (*pgxpool.Pool, error) {
	if instance == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	instance.mu.RLock()
	defer instance.mu.RUnlock()

	if instance.pool == nil {
		return nil, fmt.Errorf("database pool is nil")
	}

	return instance.pool, nil
}

func Close(ctx context.Context) error {
	if instance == nil || instance.pool == nil {
		return nil
	}

	instance.mu.Lock()
	defer instance.mu.Unlock()

	instance.pool.Close()
	instance.log.Info(ctx, "Database connection pool closed")
	return nil
}

func (m *Manager) connect(ctx context.Context) error {
	poolConfig, err := pgxpool.ParseConfig(m.cfg.GetDatabaseURL())
	if err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = int32(m.cfg.Database.MaxConns)
	poolConfig.MinConns = int32(m.cfg.Database.MinConns)
	poolConfig.MaxConnLifetime = m.cfg.Database.MaxConnLifetime
	poolConfig.MaxConnIdleTime = m.cfg.Database.MaxConnIdleTime

	// Configure connection settings
	poolConfig.ConnConfig.ConnectTimeout = 10 * time.Second
	poolConfig.ConnConfig.RuntimeParams = map[string]string{
		"application_name": "yulia-lingo",
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	m.mu.Lock()
	m.pool = pool
	m.mu.Unlock()

	m.log.Info(ctx, "Database connection pool established",
		logger.Field{Key: "max_conns", Value: m.cfg.Database.MaxConns},
		logger.Field{Key: "min_conns", Value: m.cfg.Database.MinConns},
	)

	return nil
}

func HealthCheck(ctx context.Context) error {
	pool, err := GetDB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}
