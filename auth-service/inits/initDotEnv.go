package inits

import (
	"log"

	"github.com/joho/godotenv"
)

func InitDotEnv() {
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

}
