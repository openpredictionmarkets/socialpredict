package env

import (
	"os"

	"github.com/joho/godotenv"
	"socialpredict/logger"
)

// LoadDevFile loads environment variables from .env.dev if it exists.
// It is non-fatal when the file is missing so containers can rely on
// real environment variables supplied at runtime.
func LoadDevFile() error {
	if _, err := os.Stat("./.env.dev"); os.IsNotExist(err) {
		logger.LogInfo("DevEnv", "LoadDevFile", ".env.dev not found; skipping dotenv load")
		return nil
	} else if err != nil {
		logger.LogWarn("DevEnv", "LoadDevFile", "warning checking .env.dev: "+err.Error())
	}

	if err := godotenv.Load("./.env.dev"); err != nil {
		logger.LogWarn("DevEnv", "LoadDevFile", "warning: failed to load .env.dev: "+err.Error())
	}

	return nil
}
