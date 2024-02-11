package connections

import (
	"log"

	"github.com/joho/godotenv"
)

func initEnvFile() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	log.Printf("Env file was ininzilizate secssesful")
}
