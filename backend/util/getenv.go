package util

import (
	"log"

	"github.com/joho/godotenv"
)

// GetEnv loads environment variables from the specified file
func GetEnv() error {
	err := godotenv.Load("./.env.dev")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
		return err
	}
	return nil
}
