package database

import (
	"log"

	"gorm.io/gorm"
)

// MigrateInlineRecipes adds support for inline/temporary recipes in diary entries
func MigrateInlineRecipes(db *gorm.DB) error {
	log.Println("Running inline recipes migration...")

	// Check if column already exists
	var count int64
	db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_name = 'diary_entries'
		AND column_name = 'inline_recipe_name'
	`).Scan(&count)

	if count > 0 {
		log.Println("Column inline_recipe_name already exists, skipping migration")
		return nil
	}

	// Add inline_recipe_name column
	if err := db.Exec(`
		ALTER TABLE diary_entries
		ADD COLUMN inline_recipe_name VARCHAR(255)
	`).Error; err != nil {
		return err
	}

	log.Println("Inline recipes migration completed")
	return nil
}
