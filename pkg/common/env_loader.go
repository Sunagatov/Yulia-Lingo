package common

import (
	"github.com/joho/godotenv"
	"log"
)

func Load() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}
