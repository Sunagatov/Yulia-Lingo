package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func StartHTTPServer() error {
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		return fmt.Errorf("nNo APP_PORT provided in environment variables")
	}
	log.Printf("Starting HTTP server on port %s", appPort)

	if httpServerError := http.ListenAndServe("0.0.0.0:"+appPort, nil); httpServerError != nil {
		return fmt.Errorf("failed to start HTTP server: %v", httpServerError)
	}
	log.Printf("Server running on :%s...\n", appPort)
	return nil
}
