-- Ultra API Database Initialization Script
-- This script creates all necessary tables for the Ultra application

-- Create programs table
CREATE TABLE IF NOT EXISTS programs (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create workouts table
CREATE TABLE IF NOT EXISTS workouts (
    id SERIAL PRIMARY KEY,
    program_id INTEGER REFERENCES programs(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    day_of_week INTEGER CHECK (day_of_week >= 0 AND day_of_week <= 6),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create exercises table
CREATE TABLE IF NOT EXISTS exercises (
    id SERIAL PRIMARY KEY,
    workout_id INTEGER REFERENCES workouts(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    weight DECIMAL(5,2) CHECK (weight >= 0),
    repetitions INTEGER CHECK (repetitions >= 0),
    series INTEGER CHECK (series >= 0),
    rest_time INTEGER CHECK (rest_time >= 0),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create foods table
CREATE TABLE IF NOT EXISTS foods (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    calories_per_100g INTEGER CHECK (calories_per_100g >= 0),
    protein_per_100g DECIMAL(5,2) CHECK (protein_per_100g >= 0),
    carbs_per_100g DECIMAL(5,2) CHECK (carbs_per_100g >= 0),
    fat_per_100g DECIMAL(5,2) CHECK (fat_per_100g >= 0),
    fiber_per_100g DECIMAL(5,2) CHECK (fiber_per_100g >= 0),
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create recipes table
CREATE TABLE IF NOT EXISTS recipes (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    instructions TEXT,
    servings INTEGER DEFAULT 1 CHECK (servings > 0),
    prep_time INTEGER CHECK (prep_time >= 0),
    cook_time INTEGER CHECK (cook_time >= 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create recipe_ingredients table
CREATE TABLE IF NOT EXISTS recipe_ingredients (
    id SERIAL PRIMARY KEY,
    recipe_id INTEGER REFERENCES recipes(id) ON DELETE CASCADE,
    food_id INTEGER REFERENCES foods(id),
    quantity DECIMAL(8,2) CHECK (quantity > 0),
    unit VARCHAR(50) NOT NULL
);

-- Create nutrition_goals table
CREATE TABLE IF NOT EXISTS nutrition_goals (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) UNIQUE NOT NULL,
    daily_calories INTEGER CHECK (daily_calories >= 0),
    daily_protein DECIMAL(5,2) CHECK (daily_protein >= 0),
    daily_carbs DECIMAL(5,2) CHECK (daily_carbs >= 0),
    daily_fat DECIMAL(5,2) CHECK (daily_fat >= 0),
    daily_fiber DECIMAL(5,2) CHECK (daily_fiber >= 0),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_programs_user_id ON programs(user_id);
CREATE INDEX IF NOT EXISTS idx_workouts_program_id ON workouts(program_id);
CREATE INDEX IF NOT EXISTS idx_workouts_day_of_week ON workouts(day_of_week);
CREATE INDEX IF NOT EXISTS idx_exercises_workout_id ON exercises(workout_id);
CREATE INDEX IF NOT EXISTS idx_foods_name ON foods(name);
CREATE INDEX IF NOT EXISTS idx_recipes_user_id ON recipes(user_id);
CREATE INDEX IF NOT EXISTS idx_recipe_ingredients_recipe_id ON recipe_ingredients(recipe_id);
CREATE INDEX IF NOT EXISTS idx_recipe_ingredients_food_id ON recipe_ingredients(food_id);
CREATE INDEX IF NOT EXISTS idx_nutrition_goals_user_id ON nutrition_goals(user_id);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at columns
DROP TRIGGER IF EXISTS update_programs_updated_at ON programs;
CREATE TRIGGER update_programs_updated_at
    BEFORE UPDATE ON programs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_recipes_updated_at ON recipes;
CREATE TRIGGER update_recipes_updated_at
    BEFORE UPDATE ON recipes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_nutrition_goals_updated_at ON nutrition_goals;
CREATE TRIGGER update_nutrition_goals_updated_at
    BEFORE UPDATE ON nutrition_goals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create meals table
CREATE TABLE IF NOT EXISTS meals (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    meal_type VARCHAR(50) NOT NULL CHECK (meal_type IN ('breakfast', 'lunch', 'dinner', 'snack')),
    date DATE NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create meal_items table
CREATE TABLE IF NOT EXISTS meal_items (
    id SERIAL PRIMARY KEY,
    meal_id INTEGER REFERENCES meals(id) ON DELETE CASCADE,
    item_type VARCHAR(50) NOT NULL CHECK (item_type IN ('food', 'recipe')),
    item_id INTEGER NOT NULL,
    quantity DECIMAL(8,2) NOT NULL CHECK (quantity > 0),
    notes TEXT
);

-- Create indexes for meals table
CREATE INDEX IF NOT EXISTS idx_meals_user_id ON meals(user_id);
CREATE INDEX IF NOT EXISTS idx_meals_date ON meals(date);
CREATE INDEX IF NOT EXISTS idx_meals_meal_type ON meals(meal_type);
CREATE INDEX IF NOT EXISTS idx_meals_user_date ON meals(user_id, date);

-- Create indexes for meal_items table
CREATE INDEX IF NOT EXISTS idx_meal_items_meal_id ON meal_items(meal_id);
CREATE INDEX IF NOT EXISTS idx_meal_items_item_type_id ON meal_items(item_type, item_id);

-- Create triggers for updated_at columns on meals
DROP TRIGGER IF EXISTS update_meals_updated_at ON meals;
CREATE TRIGGER update_meals_updated_at
    BEFORE UPDATE ON meals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();