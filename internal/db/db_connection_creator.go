package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose"
)

const migrationsDir = "migrations"

func CreateDatabaseConnection() *sql.DB {
	host := os.Getenv("POSTGRESQL_HOST")
	port := os.Getenv("POSTGRESQL_PORT")
	user := os.Getenv("POSTGRESQL_USER")
	password := os.Getenv("POSTGRESQL_PASSWORD")
	dbname := os.Getenv("POSTGRESQL_DATABASE_NAME")

	db := connect(host, port, user, password, dbname)
	ping(db)

	log.Println("Successfully connected to database")

	return db
}

func connect(host, port, user, password, dbname string) *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(fmt.Sprintf("Error opening database: %v", err))
	}
	return db
}

func ping(db *sql.DB) {
	// Check the connection
	err := db.Ping()
	if err != nil {
		panic(fmt.Sprintf("Error connecting to database: '%v'", err))
	}
}

func UpMigrations(db *sql.DB) {
	if err := goose.Up(db, migrationsDir); err != nil {
		panic(fmt.Sprintf("Error running migrations: %v", err))
	}
	log.Printf("Successfully run migrations")
}

func DownMigrations(db *sql.DB) {
	if err := goose.Down(db, migrationsDir); err != nil {
		panic(fmt.Sprintf("Error running migrations: %v", err))
	}
	log.Printf("Successfully run migrations")
}
