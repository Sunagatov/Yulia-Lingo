package connections

import (
	"github.com/joho/godotenv"
	"log"
)

func initEnvFile() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	log.Printf("Env file was ininzilizate secssesful")
}
