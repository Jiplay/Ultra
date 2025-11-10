package database

import (
	"encoding/json"
	"fmt"
	"log"

	"gorm.io/gorm"
)

// CustomIngredient represents an ingredient with custom quantity (for migration)
type CustomIngredient struct {
	FoodID        uint    `json:"food_id"`
	FoodName      string  `json:"food_name,omitempty"`
	QuantityGrams float64 `json:"quantity_grams"`
	Calories      float64 `json:"calories"`
	Protein       float64 `json:"protein"`
	Carbs         float64 `json:"carbs"`
	Fat           float64 `json:"fat"`
	Fiber         float64 `json:"fiber"`
}

// MigrateCustomIngredients backfills custom_ingredients JSONB field for existing recipe diary entries
// This migration:
// 1. Adds custom_ingredients JSONB column if it doesn't exist
// 2. For each recipe diary entry, calculates the proportional ingredient quantities
// 3. Populates custom_ingredients with the calculated data
func MigrateCustomIngredients(db *gorm.DB) error {
	log.Println("Starting migration to populate custom_ingredients...")

	// Start a transaction
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. Check if custom_ingredients column exists, add if not
		log.Println("Checking custom_ingredients column...")

		var hasColumn bool
		if err := tx.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'diary_entries'
				AND column_name = 'custom_ingredients'
			)
		`).Scan(&hasColumn).Error; err != nil {
			return fmt.Errorf("failed to check custom_ingredients column: %w", err)
		}

		if !hasColumn {
			// Add the column
			if err := tx.Exec(`
				ALTER TABLE diary_entries
				ADD COLUMN custom_ingredients JSONB
			`).Error; err != nil {
				return fmt.Errorf("failed to add custom_ingredients column: %w", err)
			}
			log.Println("  ✓ Added custom_ingredients JSONB column")
		} else {
			log.Println("  ✓ custom_ingredients column already exists")
		}

		// 2. Get all recipe diary entries that don't have custom_ingredients yet
		log.Println("Fetching recipe diary entries to migrate...")

		var recipeEntries []struct {
			ID            uint
			RecipeID      uint
			QuantityGrams float64
		}

		if err := tx.Raw(`
			SELECT id, recipe_id, quantity_grams
			FROM diary_entries
			WHERE recipe_id IS NOT NULL
			AND (custom_ingredients IS NULL OR custom_ingredients = 'null'::jsonb)
		`).Scan(&recipeEntries).Error; err != nil {
			return fmt.Errorf("failed to fetch recipe entries: %w", err)
		}

		log.Printf("  Found %d recipe entries to migrate", len(recipeEntries))

		// 3. For each entry, calculate custom ingredients
		for i, entry := range recipeEntries {
			// Get recipe ingredients
			var ingredients []struct {
				FoodID        uint
				FoodName      string
				QuantityGrams float64
				Calories      float64
				Protein       float64
				Carbs         float64
				Fat           float64
				Fiber         float64
			}

			if err := tx.Raw(`
				SELECT
					ri.food_id,
					f.name as food_name,
					ri.quantity_grams,
					f.calories,
					f.protein,
					f.carbs,
					f.fat,
					f.fiber
				FROM recipe_ingredients ri
				JOIN foods f ON ri.food_id = f.id
				WHERE ri.recipe_id = ?
			`, entry.RecipeID).Scan(&ingredients).Error; err != nil {
				log.Printf("  Warning: Failed to fetch ingredients for recipe %d: %v", entry.RecipeID, err)
				continue
			}

			if len(ingredients) == 0 {
				log.Printf("  Warning: Recipe %d has no ingredients, skipping entry %d", entry.RecipeID, entry.ID)
				continue
			}

			// Calculate total recipe weight
			var totalRecipeWeight float64
			for _, ing := range ingredients {
				totalRecipeWeight += ing.QuantityGrams
			}

			if totalRecipeWeight == 0 {
				log.Printf("  Warning: Recipe %d has zero weight, skipping entry %d", entry.RecipeID, entry.ID)
				continue
			}

			// Calculate portion consumed
			portion := entry.QuantityGrams / totalRecipeWeight

			// Build custom ingredients array
			var customIngredients []CustomIngredient
			for _, ing := range ingredients {
				// Calculate proportional quantity
				ingredientQuantity := ing.QuantityGrams * portion

				// Calculate nutrition for this ingredient
				multiplier := ingredientQuantity / 100.0
				customIngredients = append(customIngredients, CustomIngredient{
					FoodID:        ing.FoodID,
					FoodName:      ing.FoodName,
					QuantityGrams: roundToTwo(ingredientQuantity),
					Calories:      roundToTwo(ing.Calories * multiplier),
					Protein:       roundToTwo(ing.Protein * multiplier),
					Carbs:         roundToTwo(ing.Carbs * multiplier),
					Fat:           roundToTwo(ing.Fat * multiplier),
					Fiber:         roundToTwo(ing.Fiber * multiplier),
				})
			}

			// Convert to JSON
			customIngredientsJSON, err := json.Marshal(customIngredients)
			if err != nil {
				log.Printf("  Warning: Failed to marshal custom_ingredients for entry %d: %v", entry.ID, err)
				continue
			}

			// Update the entry
			if err := tx.Exec(`
				UPDATE diary_entries
				SET custom_ingredients = ?::jsonb
				WHERE id = ?
			`, customIngredientsJSON, entry.ID).Error; err != nil {
				log.Printf("  Warning: Failed to update entry %d: %v", entry.ID, err)
				continue
			}

			// Log progress every 100 entries
			if (i+1)%100 == 0 {
				log.Printf("  Progress: %d/%d entries migrated", i+1, len(recipeEntries))
			}
		}

		log.Printf("  ✓ Successfully migrated %d recipe entries with custom ingredients", len(recipeEntries))

		return nil
	})
}

// roundToTwo rounds a float to 2 decimal places
func roundToTwo(val float64) float64 {
	return float64(int(val*100)) / 100
}
