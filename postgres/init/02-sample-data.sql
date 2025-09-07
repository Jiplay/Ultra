-- Insert sample food data
-- This script runs after the schema is created

\c ultra_food_db;

-- Insert sample foods with nutritional information
INSERT INTO foods (name, calories, protein, carbs, fat, created_at, updated_at) VALUES
    ('Apple', 95, 0.5, 25.0, 0.3, NOW(), NOW()),
    ('Banana', 105, 1.3, 27.0, 0.4, NOW(), NOW()),
    ('Orange', 62, 1.2, 15.4, 0.2, NOW(), NOW()),
    ('Chicken Breast', 231, 43.5, 0.0, 5.0, NOW(), NOW()),
    ('Salmon Fillet', 208, 22.0, 0.0, 12.0, NOW(), NOW()),
    ('Brown Rice', 216, 5.0, 45.0, 1.8, NOW(), NOW()),
    ('Quinoa', 222, 8.1, 39.4, 3.6, NOW(), NOW()),
    ('Broccoli', 55, 3.7, 11.2, 0.6, NOW(), NOW()),
    ('Spinach', 23, 2.9, 3.6, 0.4, NOW(), NOW()),
    ('Sweet Potato', 112, 2.0, 26.0, 0.1, NOW(), NOW()),
    ('Greek Yogurt', 100, 10.0, 6.0, 5.0, NOW(), NOW()),
    ('Almonds', 576, 21.2, 21.6, 49.9, NOW(), NOW()),
    ('Oatmeal', 389, 16.9, 66.3, 6.9, NOW(), NOW()),
    ('Eggs', 155, 13.0, 1.1, 11.0, NOW(), NOW()),
    ('Avocado', 160, 2.0, 8.5, 14.7, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Create a view for high-protein foods
CREATE OR REPLACE VIEW high_protein_foods AS
SELECT id, name, calories, protein, carbs, fat, created_at
FROM foods 
WHERE protein >= 10.0
ORDER BY protein DESC;

-- Grant access to the view
GRANT SELECT ON high_protein_foods TO ultra_user;