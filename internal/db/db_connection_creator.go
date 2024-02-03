package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func CreateDatabaseConnection() (*sql.DB, error) {
	host := os.Getenv("POSTGRESQL_HOST")
	port := os.Getenv("POSTGRESQL_PORT")
	user := os.Getenv("POSTGRESQL_USER")
	password := os.Getenv("POSTGRESQL_PASSWORD")
	dbname := os.Getenv("POSTGRESQL_DATABASE_NAME")

	db, err := connect(host, port, user, password, dbname)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	err = ping(db)
	if err != nil {
		return nil, fmt.Errorf("error pinging the database: %v", err)
	}
	log.Println("Successfully connected to database")
	return db, nil
}

func connect(host, port, user, password, dbname string) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	return sql.Open("postgres", psqlInfo)
}

func ping(db *sql.DB) error {
	return db.Ping()
}
