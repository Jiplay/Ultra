package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	postgresdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDBContainer holds a test database container and its connection
type TestDBContainer struct {
	Container *postgres.PostgresContainer
	DB        *gorm.DB
	ctx       context.Context
}

// SetupTestDB creates a PostgreSQL test container and returns a GORM DB connection
// The container and connection are automatically cleaned up when the test finishes
// Note: You need to run migrations manually using AutoMigrate with your models
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	ctx := context.Background()

	// Create PostgreSQL container
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	// Clean up container when test finishes
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate PostgreSQL container: %v", err)
		}
	})

	// Get connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Connect with GORM
	db, err := gorm.Open(postgresdriver.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent mode for tests
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	return db
}


// CleanupTestData removes all data from the database (for tests that need a fresh state)
func CleanupTestData(t *testing.T, db *gorm.DB) {
	t.Helper()

	// Delete in reverse order of dependencies
	db.Exec("DELETE FROM diary_entries")
	db.Exec("DELETE FROM body_metrics")
	db.Exec("DELETE FROM nutrition_goals")
	db.Exec("DELETE FROM recipe_ingredients")
	db.Exec("DELETE FROM recipes")
	db.Exec("DELETE FROM foods")
	db.Exec("DELETE FROM users")
}

// TruncateAllTables truncates all tables (faster than DELETE for large datasets)
func TruncateAllTables(t *testing.T, db *gorm.DB) {
	t.Helper()

	db.Exec("TRUNCATE TABLE diary_entries, body_metrics, nutrition_goals, recipe_ingredients, recipes, foods, users CASCADE")
}

// WithTransaction runs a function within a database transaction and rolls it back
// Useful for isolated tests that shouldn't persist data
func WithTransaction(t *testing.T, db *gorm.DB, fn func(*gorm.DB)) {
	t.Helper()

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	fn(tx)
	tx.Rollback() // Always rollback for test isolation
}

// AssertRecordExists checks if a record exists in the database
func AssertRecordExists(t *testing.T, db *gorm.DB, model interface{}, conditions ...interface{}) {
	t.Helper()

	var count int64
	if err := db.Model(model).Where(conditions[0], conditions[1:]...).Count(&count).Error; err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}

	if count == 0 {
		t.Fatalf("Expected record to exist but found none: %+v", conditions)
	}
}

// AssertRecordNotExists checks if a record does not exist in the database
func AssertRecordNotExists(t *testing.T, db *gorm.DB, model interface{}, conditions ...interface{}) {
	t.Helper()

	var count int64
	if err := db.Model(model).Where(conditions[0], conditions[1:]...).Count(&count).Error; err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}

	if count > 0 {
		t.Fatalf("Expected no records but found %d: %+v", count, conditions)
	}
}

