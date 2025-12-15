package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"socialpredict/models"
	"socialpredict/seed"
	"socialpredict/util"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// SeedMarketsFromSQL populates the database with markets from a SQL file
func SeedMarketsFromSQL(db *gorm.DB, sqlFile string) error {
	fmt.Printf("Starting market seeding process from file: %s\n", sqlFile)

	// Check if markets already exist
	var existingCount int64
	db.Model(&models.Market{}).Count(&existingCount)

	if existingCount > 0 {
		fmt.Printf("Found %d existing markets in database.\n", existingCount)
		fmt.Print("Do you want to add markets from SQL file anyway? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborting market seeding.")
			return nil
		}
	}

	// Check if SQL file exists
	if _, err := os.Stat(sqlFile); os.IsNotExist(err) {
		return fmt.Errorf("SQL file not found: %s", sqlFile)
	}

	// Execute SQL file
	fmt.Printf("Executing SQL file: %s\n", sqlFile)
	if err := executeSQLFile(db, sqlFile); err != nil {
		return fmt.Errorf("failed to execute SQL file: %v", err)
	}

	// Count markets after insertion
	var newCount int64
	db.Model(&models.Market{}).Count(&newCount)
	fmt.Printf("Database now contains %d total markets.\n", newCount)

	return nil
}

// executeSQLFile reads and executes a SQL file
func executeSQLFile(db *gorm.DB, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var sqlBuffer strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if shouldSkipSQLLine(line) {
			continue
		}

		sqlBuffer.WriteString(line)
		sqlBuffer.WriteString(" ")

		if strings.HasSuffix(line, ";") {
			if err := executeSQLBuffer(db, &sqlBuffer); err != nil {
				return err
			}
		}
	}

	if err := executeRemainingSQL(db, &sqlBuffer); err != nil {
		return err
	}

	return scanner.Err()
}

func shouldSkipSQLLine(line string) bool {
	return line == "" || strings.HasPrefix(line, "--")
}

func executeSQLBuffer(db *gorm.DB, sqlBuffer *strings.Builder) error {
	sql := strings.TrimSpace(sqlBuffer.String())
	if sql == "" {
		sqlBuffer.Reset()
		return nil
	}
	if err := db.Exec(sql).Error; err != nil {
		return fmt.Errorf("error executing SQL: %s\nError: %v", sql, err)
	}
	sqlBuffer.Reset()
	return nil
}

func executeRemainingSQL(db *gorm.DB, sqlBuffer *strings.Builder) error {
	if sqlBuffer.Len() == 0 {
		return nil
	}
	sql := strings.TrimSpace(sqlBuffer.String())
	if sql != "" && !strings.HasSuffix(sql, ";") {
		sql += ";"
	}
	if sql != ";" {
		if err := db.Exec(sql).Error; err != nil {
			return fmt.Errorf("error executing final SQL: %s\nError: %v", sql, err)
		}
	}
	return nil
}

// loadEnvFile loads environment variables from .env file
func loadEnvFile() {
	// Get the directory where this script is located
	scriptDir, err := os.Getwd()
	if err != nil {
		log.Printf("Warning: Could not get current directory: %v", err)
		return
	}

	// Look for .env file in the parent directory (project root)
	projectRoot := filepath.Dir(scriptDir)
	envFile := filepath.Join(projectRoot, ".env")

	// Try to load .env file
	if err := godotenv.Load(envFile); err != nil {
		fmt.Printf("Warning: Could not load .env file from %s: %v\n", envFile, err)
		fmt.Println("Proceeding with existing environment variables or defaults...")
	} else {
		fmt.Printf("✓ Loaded environment variables from %s\n", envFile)
	}
}

// showUsage displays help information
func showUsage() {
	fmt.Println("Usage: go run seed_markets.go [SQL_FILE]")
	fmt.Println("")
	fmt.Println("Populates the SocialPredict database with prediction markets from a SQL file.")
	fmt.Println("")
	fmt.Println("Arguments:")
	fmt.Println("  SQL_FILE    SQL file to load (default: example_markets.sql)")
	fmt.Println("              Can be absolute path or relative to scripts directory")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run seed_markets.go                        # Use example_markets.sql")
	fmt.Println("  go run seed_markets.go example_sports_markets.sql  # Use sports markets file")
	fmt.Println("  go run seed_markets.go /path/to/custom.sql    # Use absolute path")
}

var (
	errShowUsage = errors.New("show usage")
	errAbort     = errors.New("abort seeding")
)

func main() {
	fmt.Println("SocialPredict Market Seeder")
	fmt.Println("===========================")

	sqlFile, err := parseSQLFileArg(os.Args)
	if err != nil {
		if errors.Is(err, errShowUsage) {
			showUsage()
			return
		}
		log.Fatalf("Could not parse arguments: %v", err)
	}

	sqlFile, err = resolveSQLFilePath(sqlFile)
	if err != nil {
		log.Fatalf("Could not resolve SQL file path: %v", err)
	}
	fmt.Printf("Using SQL file: %s\n", sqlFile)

	loadEnvFile()

	db, err := initDBFromEnv()
	if err != nil {
		log.Fatalf("db init: %v", err)
	}

	if err := ensureRequiredUsers(db); err != nil {
		if errors.Is(err, errAbort) {
			fmt.Println("Aborting market seeding.")
			return
		}
		log.Fatalf("User check failed: %v", err)
	}

	if err := SeedMarketsFromSQL(db, sqlFile); err != nil {
		log.Fatalf("Failed to seed markets: %v", err)
	}

	fmt.Println("\nMarket seeding completed successfully!")
}

func parseSQLFileArg(args []string) (string, error) {
	if len(args) > 1 {
		if args[1] == "-h" || args[1] == "--help" {
			return "", errShowUsage
		}
		return args[1], nil
	}
	return "example_markets.sql", nil
}

func resolveSQLFilePath(sqlFile string) (string, error) {
	if filepath.IsAbs(sqlFile) {
		return sqlFile, nil
	}
	scriptDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("Could not get current directory: %v", err)
	}
	return filepath.Join(scriptDir, sqlFile), nil
}

func initDBFromEnv() (*gorm.DB, error) {
	dbCfg, err := util.LoadDBConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("db config: %w", err)
	}

	db, err := util.InitDB(dbCfg, util.PostgresFactory{})
	if err != nil {
		return nil, fmt.Errorf("db init: %w", err)
	}

	maxAttempts := 5
	if err := seed.EnsureDBReady(db, maxAttempts); err != nil {
		return nil, fmt.Errorf("Database not ready: %w", err)
	}
	return db, nil
}

func ensureRequiredUsers(db *gorm.DB) error {
	var raischCount, markCount int64
	db.Model(&models.User{}).Where("username = ?", "raisch").Count(&raischCount)
	db.Model(&models.User{}).Where("username = ?", "mark").Count(&markCount)

	if raischCount > 0 && markCount > 0 {
		return nil
	}

	fmt.Println("Warning: Required users (raisch, mark) not found!")
	fmt.Println("The markets reference these users as creator_username.")
	fmt.Println("Please create these users first by running the main application.")
	fmt.Print("Do you want to continue anyway? (y/N): ")
	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		return errAbort
	}
	return nil
}
