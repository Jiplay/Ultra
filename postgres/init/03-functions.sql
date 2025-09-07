-- Utility functions and triggers for the Ultra Food Database

\c ultra_food_db;

-- Function to automatically update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to automatically update updated_at on foods table
DROP TRIGGER IF EXISTS update_foods_updated_at ON foods;
CREATE TRIGGER update_foods_updated_at
    BEFORE UPDATE ON foods
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to calculate BMI (example utility function)
CREATE OR REPLACE FUNCTION calculate_calories_per_gram_protein(food_id INTEGER)
RETURNS DECIMAL(10,2) AS $$
DECLARE
    food_calories INTEGER;
    food_protein DECIMAL(10,2);
    result DECIMAL(10,2);
BEGIN
    SELECT calories, protein INTO food_calories, food_protein 
    FROM foods WHERE id = food_id;
    
    IF food_protein > 0 THEN
        result := food_calories::DECIMAL / food_protein;
    ELSE
        result := 0;
    END IF;
    
    RETURN result;
END;
$$ language 'plpgsql';

-- Grant execute permissions
GRANT EXECUTE ON FUNCTION update_updated_at_column() TO ultra_user;
GRANT EXECUTE ON FUNCTION calculate_calories_per_gram_protein(INTEGER) TO ultra_user;