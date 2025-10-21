package main

import (
	"bufio"
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
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		
		sqlBuffer.WriteString(line)
		sqlBuffer.WriteString(" ")
		
		// Execute when we hit a semicolon (end of statement)
		if strings.HasSuffix(line, ";") {
			sql := strings.TrimSpace(sqlBuffer.String())
			if sql != "" {
				if err := db.Exec(sql).Error; err != nil {
					return fmt.Errorf("error executing SQL: %s\nError: %v", sql, err)
				}
			}
			sqlBuffer.Reset()
		}
	}
	
	// Execute any remaining SQL
	if sqlBuffer.Len() > 0 {
		sql := strings.TrimSpace(sqlBuffer.String())
		if sql != "" && !strings.HasSuffix(sql, ";") {
			sql += ";"
		}
		if sql != ";" {
			if err := db.Exec(sql).Error; err != nil {
				return fmt.Errorf("error executing final SQL: %s\nError: %v", sql, err)
			}
		}
	}

	return scanner.Err()
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

func main() {
	fmt.Println("SocialPredict Market Seeder")
	fmt.Println("===========================")

	// Parse command line arguments
	var sqlFile string
	if len(os.Args) > 1 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			showUsage()
			os.Exit(0)
		}
		sqlFile = os.Args[1]
	} else {
		sqlFile = "example_markets.sql"
	}

	// Convert relative path to absolute if needed
	if !filepath.IsAbs(sqlFile) {
		scriptDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Could not get current directory: %v", err)
		}
		sqlFile = filepath.Join(scriptDir, sqlFile)
	}

	fmt.Printf("Using SQL file: %s\n", sqlFile)

	// Load .env file if it exists
	loadEnvFile()

	// Initialize database connection
	util.InitDB()
	db := util.GetDB()

	// Ensure database is ready
	maxAttempts := 5
	if err := seed.EnsureDBReady(db, maxAttempts); err != nil {
		log.Fatalf("Database not ready: %v", err)
	}

	// Check if required users exist (check first few users from SQL file)
	var raischCount, markCount int64
	db.Model(&models.User{}).Where("username = ?", "raisch").Count(&raischCount)
	db.Model(&models.User{}).Where("username = ?", "mark").Count(&markCount)
	
	if raischCount == 0 || markCount == 0 {
		fmt.Println("Warning: Required users (raisch, mark) not found!")
		fmt.Println("The markets reference these users as creator_username.")
		fmt.Println("Please create these users first by running the main application.")
		fmt.Print("Do you want to continue anyway? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborting market seeding.")
			os.Exit(1)
		}
	}

	// Seed markets from SQL file
	if err := SeedMarketsFromSQL(db, sqlFile); err != nil {
		log.Fatalf("Failed to seed markets: %v", err)
	}

	fmt.Println("\nMarket seeding completed successfully!")
}