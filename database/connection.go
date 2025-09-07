package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

const (
	maxRetries = 5
	retryDelay = 2 * time.Second
)

func Connect(config *Config) (*sql.DB, error) {
	var db *sql.DB
	var err error

	connStr := config.ConnectionString()

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Failed to open database connection (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(retryDelay)
			continue
		}

		if err = db.Ping(); err != nil {
			log.Printf("Failed to ping database (attempt %d/%d): %v", i+1, maxRetries, err)
			db.Close()
			time.Sleep(retryDelay)
			continue
		}

		// Connection successful
		log.Println("Successfully connected to PostgreSQL database")
		
		// Configure connection pool
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)
		
		return db, nil
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

func HealthCheck(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Test a simple query
	var result int
	if err := db.QueryRow("SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("database query test failed: %w", err)
	}

	return nil
}