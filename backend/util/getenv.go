package util

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// GetEnv loads environment variables from .env.dev if it exists.
// It is non-fatal when the file is missing so containers can rely on
// real environment variables supplied at runtime.
func GetEnv() error {
	// If the file does not exist, skip loading silently.
	if _, err := os.Stat("./.env.dev"); os.IsNotExist(err) {
		log.Printf(".env.dev not found; skipping dotenv load")
		return nil
	} else if err != nil {
		// Unexpected stat error, log and attempt to continue
		log.Printf("Warning checking .env.dev: %v", err)
	}
  // Try to load the file; on failure, log a warning but do not return an error.
	if err := godotenv.Load("./.env.dev"); err != nil {
		log.Printf("Warning: failed to load .env.dev: %v", err)
	}
	return nil
}
