package database

import (
	"database/sql"
	"fmt"
	"log"
)

const createFoodsTable = `
CREATE TABLE IF NOT EXISTS foods (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    unit VARCHAR(10) NOT NULL DEFAULT '100g',
    calories INTEGER NOT NULL DEFAULT 0,
    protein DECIMAL(10,2) NOT NULL DEFAULT 0.0,
    carbs DECIMAL(10,2) NOT NULL DEFAULT 0.0,
    fat DECIMAL(10,2) NOT NULL DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_foods_name ON foods(name);
CREATE INDEX IF NOT EXISTS idx_foods_created_at ON foods(created_at);
`

const insertSampleData = `
INSERT INTO foods (name, unit, calories, protein, carbs, fat) VALUES
    ('Apple', 'piece', 95, 0.5, 25.0, 0.3),
    ('Banana', 'piece', 105, 1.3, 27.0, 0.4),
    ('Chicken Breast', '100g', 231, 43.5, 0.0, 5.0),
    ('Brown Rice', '100g', 216, 5.0, 45.0, 1.8),
    ('Broccoli', '100g', 55, 3.7, 11.2, 0.6)
ON CONFLICT DO NOTHING;
`

func Migrate(db *sql.DB) error {
	log.Println("Running database migrations...")

	// Create tables
	if _, err := db.Exec(createFoodsTable); err != nil {
		return fmt.Errorf("failed to create foods table: %w", err)
	}
	log.Println("✓ Foods table created/verified")

	// Insert sample data
	if _, err := db.Exec(insertSampleData); err != nil {
		return fmt.Errorf("failed to insert sample data: %w", err)
	}
	log.Println("✓ Sample data inserted/verified")

	log.Println("Database migration completed successfully")
	return nil
}