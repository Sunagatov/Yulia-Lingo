package main

import (
	"database/sql"
	"fmt"
	"log"
)

func createDatabaseConnection(host string, port string, user string, password string, dbname string) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// Check the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to database: '%v'", err)
	}
	log.Println("Successfully connected to database")
	return db, err
}
