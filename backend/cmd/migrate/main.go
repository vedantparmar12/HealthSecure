package main

import (
	"fmt"
	"log"
	"os"

	"healthsecure/configs"
	"healthsecure/internal/database"
)

func main() {
	fmt.Println("HealthSecure Database Migration Tool")
	fmt.Println("===================================")

	// Load configuration
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	if err := database.Initialize(config); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Check command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "up":
			runMigrations()
		case "status":
			checkMigrationStatus()
		case "clean":
			runCleanupTasks()
		default:
			printUsage()
		}
	} else {
		runMigrations()
	}
}

func runMigrations() {
	log.Println("Running database migrations...")
	
	// The migrations are automatically run during database.Initialize()
	// This is just for manual migration runs
	
	log.Println("✅ Database migrations completed successfully")
}

func checkMigrationStatus() {
	log.Println("Checking database connection and schema...")
	
	if err := database.Health(); err != nil {
		log.Fatalf("❌ Database health check failed: %v", err)
	}
	
	log.Println("✅ Database is healthy and accessible")
}

func runCleanupTasks() {
	log.Println("Running database cleanup tasks...")
	
	if err := database.RunCleanupTasks(); err != nil {
		log.Fatalf("❌ Cleanup tasks failed: %v", err)
	}
	
	log.Println("✅ Cleanup tasks completed successfully")
}

func printUsage() {
	fmt.Println("Usage: go run cmd/migrate/main.go [command]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  up      Run database migrations (default)")
	fmt.Println("  status  Check database status and connectivity")
	fmt.Println("  clean   Run database cleanup tasks")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/migrate/main.go")
	fmt.Println("  go run cmd/migrate/main.go up")
	fmt.Println("  go run cmd/migrate/main.go status")
	fmt.Println("  go run cmd/migrate/main.go clean")
}