package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
)

var (
	postgres *sql.DB
	err      error
)

func CreateDatabaseConnection() {
	host := os.Getenv("POSTGRESQL_HOST")
	port := os.Getenv("POSTGRESQL_PORT")
	user := os.Getenv("POSTGRESQL_USER")
	password := os.Getenv("POSTGRESQL_PASSWORD")
	dbname := os.Getenv("POSTGRESQL_DATABASE_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	postgres, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	if err = postgres.Ping(); err != nil {
		log.Fatalf("error pinging the database: %v", err)
	}

	log.Println("Successfully connected to database")
}

func CloseDatabaseConnection() {
	postgres.Close()
}

func GetPostgresClient() (*sql.DB, error) {
	return postgres, err
}
