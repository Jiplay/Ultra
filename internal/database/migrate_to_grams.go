package database

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// MigrateToGrams performs the migration from serving-based to gram-based quantities
// This migration:
// 1. Renames columns: quantity -> quantity_grams, serving_size -> quantity_grams
// 2. Multiplies existing values by 100 (assuming 1 serving = 100g)
// 3. Removes the recipes.serving_size column
// 4. Recalculates all cached nutrition values in diary entries
func MigrateToGrams(db *gorm.DB) error {
	log.Println("Starting migration to gram-based system...")

	// Start a transaction
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. Migrate recipe_ingredients table
		log.Println("Migrating recipe_ingredients table...")

		// Check if old column exists
		var hasOldColumn bool
		if err := tx.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'recipe_ingredients'
				AND column_name = 'quantity'
			)
		`).Scan(&hasOldColumn).Error; err != nil {
			return fmt.Errorf("failed to check recipe_ingredients columns: %w", err)
		}

		if hasOldColumn {
			// Rename column and multiply values by 100
			if err := tx.Exec(`
				ALTER TABLE recipe_ingredients
				RENAME COLUMN quantity TO quantity_grams
			`).Error; err != nil {
				return fmt.Errorf("failed to rename recipe_ingredients.quantity: %w", err)
			}

			if err := tx.Exec(`
				UPDATE recipe_ingredients
				SET quantity_grams = quantity_grams * 100
			`).Error; err != nil {
				return fmt.Errorf("failed to update recipe_ingredients quantities: %w", err)
			}

			log.Println("  ✓ Migrated recipe_ingredients.quantity -> quantity_grams (* 100)")
		} else {
			log.Println("  ✓ recipe_ingredients already migrated")
		}

		// 2. Migrate diary_entries table
		log.Println("Migrating diary_entries table...")

		var hasDiaryOldColumn bool
		if err := tx.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'diary_entries'
				AND column_name = 'serving_size'
			)
		`).Scan(&hasDiaryOldColumn).Error; err != nil {
			return fmt.Errorf("failed to check diary_entries columns: %w", err)
		}

		if hasDiaryOldColumn {
			// Rename column and multiply values by 100
			if err := tx.Exec(`
				ALTER TABLE diary_entries
				RENAME COLUMN serving_size TO quantity_grams
			`).Error; err != nil {
				return fmt.Errorf("failed to rename diary_entries.serving_size: %w", err)
			}

			if err := tx.Exec(`
				UPDATE diary_entries
				SET quantity_grams = quantity_grams * 100
			`).Error; err != nil {
				return fmt.Errorf("failed to update diary_entries quantities: %w", err)
			}

			log.Println("  ✓ Migrated diary_entries.serving_size -> quantity_grams (* 100)")
		} else {
			log.Println("  ✓ diary_entries already migrated")
		}

		// 3. Remove recipes.serving_size column if it exists
		log.Println("Removing recipes.serving_size column...")

		var hasRecipeServingSize bool
		if err := tx.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'recipes'
				AND column_name = 'serving_size'
			)
		`).Scan(&hasRecipeServingSize).Error; err != nil {
			return fmt.Errorf("failed to check recipes columns: %w", err)
		}

		if hasRecipeServingSize {
			if err := tx.Exec(`
				ALTER TABLE recipes
				DROP COLUMN serving_size
			`).Error; err != nil {
				return fmt.Errorf("failed to drop recipes.serving_size: %w", err)
			}

			log.Println("  ✓ Removed recipes.serving_size column")
		} else {
			log.Println("  ✓ recipes.serving_size already removed")
		}

		// 4. Recalculate diary entry nutrition values
		// We need to recalculate because the formula has changed
		log.Println("Recalculating diary entry nutrition values...")

		// For food-based entries: nutrition = food_nutrition * (quantity_grams / 100)
		if err := tx.Exec(`
			UPDATE diary_entries de
			SET
				calories = f.calories * (de.quantity_grams / 100.0),
				protein = f.protein * (de.quantity_grams / 100.0),
				carbs = f.carbs * (de.quantity_grams / 100.0),
				fat = f.fat * (de.quantity_grams / 100.0),
				fiber = f.fiber * (de.quantity_grams / 100.0)
			FROM foods f
			WHERE de.food_id = f.id
		`).Error; err != nil {
			return fmt.Errorf("failed to recalculate food-based diary entries: %w", err)
		}

		// For recipe-based entries, we need to:
		// 1. Calculate total recipe weight
		// 2. Calculate total recipe nutrition
		// 3. Apply the portion consumed
		// This is complex, so we'll do it in Go
		var recipeEntries []struct {
			ID       uint
			RecipeID uint
			QuantityGrams float64
		}

		if err := tx.Raw(`
			SELECT id, recipe_id, quantity_grams
			FROM diary_entries
			WHERE recipe_id IS NOT NULL
		`).Scan(&recipeEntries).Error; err != nil {
			return fmt.Errorf("failed to fetch recipe entries: %w", err)
		}

		for _, entry := range recipeEntries {
			// Calculate total recipe nutrition
			var nutrition struct {
				TotalCalories float64
				TotalProtein  float64
				TotalCarbs    float64
				TotalFat      float64
				TotalFiber    float64
				TotalWeight   float64
			}

			if err := tx.Raw(`
				SELECT
					SUM(f.calories * (ri.quantity_grams / 100.0)) as total_calories,
					SUM(f.protein * (ri.quantity_grams / 100.0)) as total_protein,
					SUM(f.carbs * (ri.quantity_grams / 100.0)) as total_carbs,
					SUM(f.fat * (ri.quantity_grams / 100.0)) as total_fat,
					SUM(f.fiber * (ri.quantity_grams / 100.0)) as total_fiber,
					SUM(ri.quantity_grams) as total_weight
				FROM recipe_ingredients ri
				JOIN foods f ON ri.food_id = f.id
				WHERE ri.recipe_id = ?
			`, entry.RecipeID).Scan(&nutrition).Error; err != nil {
				return fmt.Errorf("failed to calculate recipe %d nutrition: %w", entry.RecipeID, err)
			}

			// Calculate portion: consumed_grams / total_recipe_grams
			if nutrition.TotalWeight > 0 {
				portion := entry.QuantityGrams / nutrition.TotalWeight

				if err := tx.Exec(`
					UPDATE diary_entries
					SET
						calories = ?,
						protein = ?,
						carbs = ?,
						fat = ?,
						fiber = ?
					WHERE id = ?
				`,
					nutrition.TotalCalories * portion,
					nutrition.TotalProtein * portion,
					nutrition.TotalCarbs * portion,
					nutrition.TotalFat * portion,
					nutrition.TotalFiber * portion,
					entry.ID,
				).Error; err != nil {
					return fmt.Errorf("failed to update diary entry %d: %w", entry.ID, err)
				}
			}
		}

		log.Println("  ✓ Recalculated nutrition for all diary entries")

		// Log summary
		var stats struct {
			RecipeIngredients int64
			DiaryEntries      int64
		}
		tx.Raw("SELECT COUNT(*) FROM recipe_ingredients").Scan(&stats.RecipeIngredients)
		tx.Raw("SELECT COUNT(*) FROM diary_entries").Scan(&stats.DiaryEntries)

		log.Printf("Migration completed successfully!")
		log.Printf("  - Migrated %d recipe ingredients", stats.RecipeIngredients)
		log.Printf("  - Migrated %d diary entries", stats.DiaryEntries)

		return nil
	})
}
