package common

import (
	"fmt"

	"github.com/joho/godotenv"
)

func initEnvFile() {
	err := godotenv.Load()
	if err != nil {
		panic(fmt.Sprintf("Error loading .env file: %v", err))
	}
}
