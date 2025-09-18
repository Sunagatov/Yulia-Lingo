package database

import (
	"database/sql"
	"fmt"
	"sync"

	"Yulia-Lingo/internal/config"
	"Yulia-Lingo/internal/logger"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Manager struct {
	db   *sql.DB
	mu   sync.RWMutex
	cfg  *config.Config
}

var instance *Manager
var once sync.Once

func Initialize(cfg *config.Config) error {
	var err error
	once.Do(func() {
		instance = &Manager{cfg: cfg}
		err = instance.connect()
	})
	return err
}

func GetDB() (*sql.DB, error) {
	if instance == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	
	instance.mu.RLock()
	defer instance.mu.RUnlock()
	
	if instance.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	
	if err := instance.db.Ping(); err != nil {
		logger.Warn("Database ping failed, attempting reconnection", logrus.Fields{"error": err.Error()})
		if reconnectErr := instance.reconnect(); reconnectErr != nil {
			return nil, fmt.Errorf("failed to reconnect to database: %w", reconnectErr)
		}
	}
	
	return instance.db, nil
}

func Close() error {
	if instance == nil || instance.db == nil {
		return nil
	}
	
	instance.mu.Lock()
	defer instance.mu.Unlock()
	
	if err := instance.db.Close(); err != nil {
		logger.Error("Failed to close database connection", err)
		return err
	}
	
	instance.db = nil
	return nil
}

func (m *Manager) connect() error {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		m.cfg.Database.Host,
		m.cfg.Database.Port,
		m.cfg.Database.User,
		m.cfg.Database.Password,
		m.cfg.Database.Name)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	m.mu.Lock()
	m.db = db
	m.mu.Unlock()

	return nil
}

func (m *Manager) reconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.db != nil {
		m.db.Close()
	}
	
	return m.connect()
}