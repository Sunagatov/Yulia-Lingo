package server

import (
	"fmt"
	"log"
	"net/http"
)

func StartHTTPServer() error {
	appPort := "8083"
	if appPort == "" {
		return fmt.Errorf("no APP_PORT provided in environment variables")
	}
	log.Printf("Starting HTTP server on port %s", appPort)

	err := http.ListenAndServe("0.0.0.0:"+appPort, nil)
	if err != nil {
		return fmt.Errorf("failed to start HTTP server: %v", err)
	}
	log.Printf("Server running on :%s...\n", appPort)
	return nil
}
