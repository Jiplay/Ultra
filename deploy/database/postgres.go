package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var PostgresDB *sql.DB

func ConnectPostgres(host, port, user, password, dbname, sslmode string) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	var err error
	PostgresDB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	if err = PostgresDB.Ping(); err != nil {
		log.Fatal("PostgreSQL ping failed:", err)
	}

	log.Println("Connected to PostgreSQL")
}

func InitPostgresTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS programs (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS workouts (
			id SERIAL PRIMARY KEY,
			program_id INTEGER REFERENCES programs(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			day_of_week INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS exercises (
			id SERIAL PRIMARY KEY,
			workout_id INTEGER REFERENCES workouts(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			weight DECIMAL(5,2),
			repetitions INTEGER,
			series INTEGER,
			rest_time INTEGER,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS foods (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			calories_per_100g INTEGER,
			protein_per_100g DECIMAL(5,2),
			carbs_per_100g DECIMAL(5,2),
			fat_per_100g DECIMAL(5,2),
			fiber_per_100g DECIMAL(5,2),
			created_by VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS recipes (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			instructions TEXT,
			servings INTEGER DEFAULT 1,
			prep_time INTEGER,
			cook_time INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS recipe_ingredients (
			id SERIAL PRIMARY KEY,
			recipe_id INTEGER REFERENCES recipes(id) ON DELETE CASCADE,
			food_id INTEGER REFERENCES foods(id),
			quantity DECIMAL(8,2),
			unit VARCHAR(50)
		)`,
		
		`CREATE TABLE IF NOT EXISTS nutrition_goals (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(255) UNIQUE NOT NULL,
			daily_calories INTEGER,
			daily_protein DECIMAL(5,2),
			daily_carbs DECIMAL(5,2),
			daily_fat DECIMAL(5,2),
			daily_fiber DECIMAL(5,2),
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := PostgresDB.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	log.Println("PostgreSQL tables initialized")
	return nil
}