package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"os"
)

var (
	databaseConnection *sql.DB
	errors             error
)

func CreateDatabaseConnection() error {
	host := os.Getenv("POSTGRESQL_HOST")
	port := os.Getenv("POSTGRESQL_PORT")
	user := os.Getenv("POSTGRESQL_USER")
	password := os.Getenv("POSTGRESQL_PASSWORD")
	dbname := os.Getenv("POSTGRESQL_DATABASE_NAME")
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	databaseConnection, errors = sql.Open("postgres", psqlInfo)
	if errors != nil {
		return fmt.Errorf("failed to connect to database: %v", errors)
	}
	if errors = databaseConnection.Ping(); errors != nil {
		return fmt.Errorf("failed to ping the postgres database: %v", errors)
	}
	return nil
}

func CloseDatabaseConnection() {
	databaseConnection.Close()
}

func GetPostgresClient() (*sql.DB, error) {
	return databaseConnection, errors
}
