package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database  DatabaseConfig
	Telegram  TelegramConfig
	Translate TranslateConfig
	Logging   LoggingConfig
	App       AppConfig
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	MaxConns        int
	MinConns        int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	SSLMode         string
}

type TelegramConfig struct {
	BotToken           string
	MaxConcurrentUsers int
	Timeout            time.Duration
}

type TranslateConfig struct {
	APIURL  string
	APIKey  string
	APIHost string
	Timeout time.Duration
}

type LoggingConfig struct {
	Level  string
	Format string
}

type AppConfig struct {
	IrregularVerbsFilePath string
	GracefulShutdownTime   time.Duration
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	config := &Config{
		Database: DatabaseConfig{
			Host:            getEnv("POSTGRESQL_HOST", "localhost"),
			Port:            getEnvInt("POSTGRESQL_PORT", 5432),
			User:            getEnv("POSTGRESQL_USER", "postgres"),
			Password:        getEnv("POSTGRESQL_PASSWORD", ""),
			Name:            getEnv("POSTGRESQL_DATABASE_NAME", "yulia_lingo"),
			MaxConns:        getEnvInt("DB_MAX_CONNS", 25),
			MinConns:        getEnvInt("DB_MIN_CONNS", 5),
			MaxConnLifetime: getEnvDuration("DB_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime: getEnvDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
		},
		Telegram: TelegramConfig{
			BotToken:           getEnv("TELEGRAM_BOT_TOKEN", ""),
			MaxConcurrentUsers: getEnvInt("TELEGRAM_MAX_CONCURRENT_USERS", 100),
			Timeout:            getEnvDuration("TELEGRAM_TIMEOUT", 60*time.Second),
		},
		Translate: TranslateConfig{
			APIURL:  getEnv("TRANSLATE_API_URL", ""),
			APIKey:  getEnv("TRANSLATE_API_KEY", ""),
			APIHost: getEnv("TRANSLATE_API_HOST", ""),
			Timeout: getEnvDuration("TRANSLATE_API_TIMEOUT", 10*time.Second),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		App: AppConfig{
			IrregularVerbsFilePath: getEnv("IRREGULAR_VERBS_FILE_PATH", "resource/nepravilnye-glagoly-295.xlsx"),
			GracefulShutdownTime:   getEnvDuration("GRACEFUL_SHUTDOWN_TIME", 30*time.Second),
		},
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

func (c *Config) validate() error {
	if c.Telegram.BotToken == "" {
		return fmt.Errorf("missing required: TELEGRAM_BOT_TOKEN")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("missing required: POSTGRESQL_PASSWORD")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", c.Database.Port)
	}
	if c.Database.MaxConns < c.Database.MinConns {
		return fmt.Errorf("max_conns (%d) < min_conns (%d)", c.Database.MaxConns, c.Database.MinConns)
	}
	return nil
}

func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User, c.Database.Password,
		c.Database.Host, c.Database.Port,
		c.Database.Name, c.Database.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}