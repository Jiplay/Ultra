package database

import (
	"log"

	"gorm.io/gorm"
)

// MigrateRemoveMetricsConstraint removes the unique constraint on body_metrics table
func MigrateRemoveMetricsConstraint(db *gorm.DB) error {
	log.Println("Removing unique constraint from body_metrics table...")

	// List all constraints on body_metrics table
	var constraints []struct {
		ConstraintName string
		ConstraintType string
	}

	if err := db.Raw(`
		SELECT conname as constraint_name,
		       CASE contype
		           WHEN 'u' THEN 'UNIQUE'
		           WHEN 'p' THEN 'PRIMARY KEY'
		           WHEN 'f' THEN 'FOREIGN KEY'
		           WHEN 'c' THEN 'CHECK'
		       END as constraint_type
		FROM pg_constraint
		JOIN pg_class ON pg_constraint.conrelid = pg_class.oid
		WHERE pg_class.relname = 'body_metrics'
	`).Scan(&constraints).Error; err != nil {
		log.Printf("Failed to list constraints: %v", err)
		return err
	}

	log.Printf("Found %d constraints on body_metrics:", len(constraints))
	for _, c := range constraints {
		log.Printf("  - %s (%s)", c.ConstraintName, c.ConstraintType)
	}

	// Drop the unique constraint if it exists
	if err := db.Exec(`
		ALTER TABLE body_metrics
		DROP CONSTRAINT IF EXISTS idx_user_date_unique
	`).Error; err != nil {
		log.Printf("Failed to drop constraint idx_user_date_unique: %v", err)
		return err
	}

	// Also drop the unique index if it exists
	if err := db.Exec(`
		DROP INDEX IF EXISTS idx_user_date_unique
	`).Error; err != nil {
		log.Printf("Failed to drop index idx_user_date_unique: %v", err)
		return err
	}

	log.Println("  ✓ Dropped constraint and index idx_user_date_unique (if existed)")

	return nil
}
