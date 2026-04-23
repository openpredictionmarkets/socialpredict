package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadDevFile loads environment variables from .env.dev if it exists.
// It is non-fatal when the file is missing so containers can rely on
// real environment variables supplied at runtime.
func LoadDevFile() error {
	if _, err := os.Stat("./.env.dev"); os.IsNotExist(err) {
		log.Printf(".env.dev not found; skipping dotenv load")
		return nil
	} else if err != nil {
		log.Printf("Warning checking .env.dev: %v", err)
	}

	if err := godotenv.Load("./.env.dev"); err != nil {
		log.Printf("Warning: failed to load .env.dev: %v", err)
	}

	return nil
}
