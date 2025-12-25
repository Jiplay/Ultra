package database

import (
	"log"

	"gorm.io/gorm"
)

// MigrateInlineFoods adds support for inline/temporary foods in diary entries
func MigrateInlineFoods(db *gorm.DB) error {
	log.Println("Running inline foods migration...")

	// Check if column already exists
	var count int64
	db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_name = 'diary_entries'
		AND column_name = 'inline_food_name'
	`).Scan(&count)

	if count > 0 {
		log.Println("Column inline_food_name already exists, skipping migration")
		return nil
	}

	// Add all inline food columns in a single ALTER TABLE statement
	if err := db.Exec(`
		ALTER TABLE diary_entries
		ADD COLUMN inline_food_name VARCHAR(255),
		ADD COLUMN inline_food_description TEXT,
		ADD COLUMN inline_food_calories DECIMAL(10,2),
		ADD COLUMN inline_food_protein DECIMAL(10,2),
		ADD COLUMN inline_food_carbs DECIMAL(10,2),
		ADD COLUMN inline_food_fat DECIMAL(10,2),
		ADD COLUMN inline_food_fiber DECIMAL(10,2),
		ADD COLUMN inline_food_tag VARCHAR(20)
	`).Error; err != nil {
		return err
	}

	log.Println("Inline foods migration completed")
	return nil
}
