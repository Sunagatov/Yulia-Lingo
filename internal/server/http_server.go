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
		return fmt.Errorf("no APP_PORT provided in environment variables")
	}
	log.Printf("Starting HTTP server on port %s", appPort)

	go func() {
		err := http.ListenAndServe("0.0.0.0:"+appPort, nil)
		if err != nil {
			log.Fatalf("failed to start HTTP server: %v", err)
		}
	}()

	log.Printf("Server running on :%s...\n", appPort)
	return nil
}
