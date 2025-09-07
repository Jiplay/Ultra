-- Initialize Ultra Food Database
-- This script runs automatically when the PostgreSQL container starts

-- Ensure we're using the correct database
\c ultra_food_db;

-- Create foods table
CREATE TABLE IF NOT EXISTS foods (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    calories INTEGER NOT NULL DEFAULT 0,
    protein DECIMAL(10,2) NOT NULL DEFAULT 0.0,
    carbs DECIMAL(10,2) NOT NULL DEFAULT 0.0,
    fat DECIMAL(10,2) NOT NULL DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_foods_name ON foods(name);
CREATE INDEX IF NOT EXISTS idx_foods_created_at ON foods(created_at);
CREATE INDEX IF NOT EXISTS idx_foods_calories ON foods(calories);

-- Grant permissions to ultra_user
GRANT ALL PRIVILEGES ON TABLE foods TO ultra_user;
GRANT USAGE, SELECT ON SEQUENCE foods_id_seq TO ultra_user;