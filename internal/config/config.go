package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Database  DatabaseConfig
	Telegram  TelegramConfig
	Translate TranslateConfig
	Server    ServerConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type TelegramConfig struct {
	BotToken string
}

type TranslateConfig struct {
	APIURL  string
	APIKey  string
	APIHost string
}

type ServerConfig struct {
	Port string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRESQL_HOST", "localhost"),
			Port:     getEnv("POSTGRESQL_PORT", "5432"),
			User:     getEnv("POSTGRESQL_USER", "postgres"),
			Password: getEnv("POSTGRESQL_PASSWORD", ""),
			Name:     getEnv("POSTGRESQL_DATABASE_NAME", "yulia_lingo"),
		},
		Telegram: TelegramConfig{
			BotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		},
		Translate: TranslateConfig{
			APIURL:  getEnv("YOUR_TRANSLATE_API_URL", ""),
			APIKey:  getEnv("YOUR_TRANSLATE_API_KEY", ""),
			APIHost: getEnv("YOUR_TRANSLATE_API_HOST", ""),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8083"),
		},
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

func (c *Config) validate() error {
	var missing []string

	if c.Telegram.BotToken == "" {
		missing = append(missing, "TELEGRAM_BOT_TOKEN")
	}
	if c.Database.Password == "" {
		missing = append(missing, "POSTGRESQL_PASSWORD")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

func (c *Config) GetIrregularVerbsFilePath() (string, error) {
	filePath := getEnv("IRREGULAR_VERBS_FILE_PATH", "")
	if filePath == "" {
		filePath = filepath.Join("resource", "nepravilnye-glagoly-295.xlsx")
	}

	if !filepath.IsAbs(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve absolute path for irregular verbs file: %w", err)
		}
		filePath = absPath
	}

	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("irregular verbs file not found at path: %s", filePath)
		}
		return "", fmt.Errorf("failed to access irregular verbs file: %w", err)
	}

	return filePath, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}