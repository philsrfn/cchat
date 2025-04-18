package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// DB is the global database connection
var DB *sql.DB

// Config holds database configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DefaultConfig returns a default database configuration
func DefaultConfig() Config {
	return Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "phil"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "gotext"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// Init initializes the database connection
func Init(config Config) error {
	// First connect to postgres to check connectivity
	tempConnStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.SSLMode,
	)
	
	tempDB, tempErr := sql.Open("postgres", tempConnStr)
	if tempErr == nil {
		// Test connection and close temporary connection
		tempDB.Ping()
		tempDB.Close()
	}

	// Now connect to the specific database
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	// Add debug logging
	fmt.Printf("Connecting with: %s\n", connStr)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
} 