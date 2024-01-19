package main

import (
	"log"
	"net/http"
	"os"
)

func startHTTPServer() {
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		log.Fatal("No APP_PORT provided in environment variables")
	}
	log.Printf("Starting HTTP server on port %s", appPort)
	go func() {
		if httpServerError := http.ListenAndServe("0.0.0.0:"+appPort, nil); httpServerError != nil {
			log.Fatalf("Failed to start HTTP server: %v", httpServerError)
		}
	}()
}
