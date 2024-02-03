package main

import (
	database "Yulia-Lingo/internal/db"
	"Yulia-Lingo/internal/server"
	"fmt"
	_ "github.com/lib/pq"
	"net/http"
)

func main() {
	dbConnection, err := database.CreateDatabaseConnection()
	if err != nil {
		panic(fmt.Sprintf("Error opening database: %v", err))
	}

	err = database.InitDatabase(dbConnection)
	if err != nil {
		panic(fmt.Sprintf("Error opening database: %v", err))
	}

	http.HandleFunc("/irregular-verbs", database.GetIrregularVerbs)
	err = server.StartHTTPServer()
	if err != nil {
		panic(fmt.Sprintf("Error starting HTTP server: %v", err))
	}
}
