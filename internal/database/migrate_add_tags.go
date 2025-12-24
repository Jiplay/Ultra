package database

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// MigrateAddTags adds tag columns to foods, recipes, and diary_entries tables
// This migration:
// 1. Adds tag column to foods table with default 'routine'
// 2. Adds tag column to recipes table with default 'routine'
// 3. Adds food_tag and recipe_tag columns to diary_entries table
// 4. Adds check constraints to ensure only valid tag values
// 5. Backfills existing records with 'routine' tag
func MigrateAddTags(db *gorm.DB) error {
	log.Println("Starting migration to add tag columns...")

	// Start a transaction
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. Add tag column to foods table
		log.Println("Adding tag column to foods table...")

		var hasFoodTag bool
		if err := tx.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'foods'
				AND column_name = 'tag'
			)
		`).Scan(&hasFoodTag).Error; err != nil {
			return fmt.Errorf("failed to check foods.tag column: %w", err)
		}

		if !hasFoodTag {
			// Add tag column with default value
			if err := tx.Exec(`
				ALTER TABLE foods
				ADD COLUMN tag VARCHAR(20) NOT NULL DEFAULT 'routine'
			`).Error; err != nil {
				return fmt.Errorf("failed to add foods.tag column: %w", err)
			}

			// Add check constraint
			if err := tx.Exec(`
				ALTER TABLE foods
				ADD CONSTRAINT foods_tag_check CHECK (tag IN ('routine', 'contextual'))
			`).Error; err != nil {
				return fmt.Errorf("failed to add foods.tag check constraint: %w", err)
			}

			// Backfill existing records (already done by DEFAULT, but ensuring)
			if err := tx.Exec(`
				UPDATE foods
				SET tag = 'routine'
				WHERE tag IS NULL
			`).Error; err != nil {
				return fmt.Errorf("failed to backfill foods.tag: %w", err)
			}

			log.Println("  ✓ Added tag column to foods table")
		} else {
			log.Println("  ✓ foods.tag column already exists")
		}

		// 2. Add tag column to recipes table
		log.Println("Adding tag column to recipes table...")

		var hasRecipeTag bool
		if err := tx.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'recipes'
				AND column_name = 'tag'
			)
		`).Scan(&hasRecipeTag).Error; err != nil {
			return fmt.Errorf("failed to check recipes.tag column: %w", err)
		}

		if !hasRecipeTag {
			// Add tag column with default value
			if err := tx.Exec(`
				ALTER TABLE recipes
				ADD COLUMN tag VARCHAR(20) NOT NULL DEFAULT 'routine'
			`).Error; err != nil {
				return fmt.Errorf("failed to add recipes.tag column: %w", err)
			}

			// Add check constraint
			if err := tx.Exec(`
				ALTER TABLE recipes
				ADD CONSTRAINT recipes_tag_check CHECK (tag IN ('routine', 'contextual'))
			`).Error; err != nil {
				return fmt.Errorf("failed to add recipes.tag check constraint: %w", err)
			}

			// Backfill existing records
			if err := tx.Exec(`
				UPDATE recipes
				SET tag = 'routine'
				WHERE tag IS NULL
			`).Error; err != nil {
				return fmt.Errorf("failed to backfill recipes.tag: %w", err)
			}

			log.Println("  ✓ Added tag column to recipes table")
		} else {
			log.Println("  ✓ recipes.tag column already exists")
		}

		// 3. Add food_tag and recipe_tag columns to diary_entries table
		log.Println("Adding tag columns to diary_entries table...")

		var hasFoodTagColumn bool
		if err := tx.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'diary_entries'
				AND column_name = 'food_tag'
			)
		`).Scan(&hasFoodTagColumn).Error; err != nil {
			return fmt.Errorf("failed to check diary_entries.food_tag column: %w", err)
		}

		if !hasFoodTagColumn {
			// Add food_tag column (nullable - only populated for food entries)
			if err := tx.Exec(`
				ALTER TABLE diary_entries
				ADD COLUMN food_tag VARCHAR(20)
			`).Error; err != nil {
				return fmt.Errorf("failed to add diary_entries.food_tag column: %w", err)
			}

			log.Println("  ✓ Added food_tag column to diary_entries table")
		} else {
			log.Println("  ✓ diary_entries.food_tag column already exists")
		}

		var hasRecipeTagColumn bool
		if err := tx.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'diary_entries'
				AND column_name = 'recipe_tag'
			)
		`).Scan(&hasRecipeTagColumn).Error; err != nil {
			return fmt.Errorf("failed to check diary_entries.recipe_tag column: %w", err)
		}

		if !hasRecipeTagColumn {
			// Add recipe_tag column (nullable - only populated for recipe entries)
			if err := tx.Exec(`
				ALTER TABLE diary_entries
				ADD COLUMN recipe_tag VARCHAR(20)
			`).Error; err != nil {
				return fmt.Errorf("failed to add diary_entries.recipe_tag column: %w", err)
			}

			log.Println("  ✓ Added recipe_tag column to diary_entries table")
		} else {
			log.Println("  ✓ diary_entries.recipe_tag column already exists")
		}

		// 4. Backfill diary_entries with tags from foods and recipes
		log.Println("Backfilling diary entries with tags...")

		// Backfill food tags
		if err := tx.Exec(`
			UPDATE diary_entries de
			SET food_tag = f.tag
			FROM foods f
			WHERE de.food_id = f.id
			AND de.food_tag IS NULL
		`).Error; err != nil {
			return fmt.Errorf("failed to backfill diary_entries.food_tag: %w", err)
		}

		// Backfill recipe tags
		if err := tx.Exec(`
			UPDATE diary_entries de
			SET recipe_tag = r.tag
			FROM recipes r
			WHERE de.recipe_id = r.id
			AND de.recipe_tag IS NULL
		`).Error; err != nil {
			return fmt.Errorf("failed to backfill diary_entries.recipe_tag: %w", err)
		}

		log.Println("  ✓ Backfilled diary entries with tags")

		// Log summary
		var stats struct {
			Foods        int64
			Recipes      int64
			DiaryEntries int64
		}
		tx.Raw("SELECT COUNT(*) FROM foods").Scan(&stats.Foods)
		tx.Raw("SELECT COUNT(*) FROM recipes").Scan(&stats.Recipes)
		tx.Raw("SELECT COUNT(*) FROM diary_entries").Scan(&stats.DiaryEntries)

		log.Printf("Migration completed successfully!")
		log.Printf("  - Tagged %d foods", stats.Foods)
		log.Printf("  - Tagged %d recipes", stats.Recipes)
		log.Printf("  - Updated %d diary entries", stats.DiaryEntries)

		return nil
	})
}
