-- Sample data for Ultra API development
-- This script populates the database with sample data for testing

-- Insert sample foods
INSERT INTO foods (name, calories_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g, fiber_per_100g, created_by) VALUES
('Chicken Breast', 165, 31.0, 0.0, 3.6, 0.0, 'system'),
('Brown Rice', 123, 2.6, 23.0, 0.9, 1.8, 'system'),
('Broccoli', 34, 2.8, 7.0, 0.4, 2.6, 'system'),
('Salmon', 208, 20.4, 0.0, 13.4, 0.0, 'system'),
('Sweet Potato', 86, 1.6, 20.1, 0.1, 3.0, 'system'),
('Spinach', 23, 2.9, 3.6, 0.4, 2.2, 'system'),
('Banana', 89, 1.1, 23.0, 0.3, 2.6, 'system'),
('Oats', 389, 16.9, 66.3, 6.9, 10.6, 'system'),
('Greek Yogurt', 59, 10.0, 3.6, 0.4, 0.0, 'system'),
('Almonds', 579, 21.2, 21.6, 49.9, 12.5, 'system')
ON CONFLICT DO NOTHING;

-- Insert sample program
INSERT INTO programs (user_id, name, description) VALUES 
('demo-user', 'Full Body Strength', 'A comprehensive 3-day full body workout program')
ON CONFLICT DO NOTHING;

-- Get the program ID (assuming it's 1 for the first insert)
-- Insert sample workouts
INSERT INTO workouts (program_id, name, description, day_of_week) VALUES
(1, 'Upper Body Focus', 'Focus on chest, back, shoulders and arms', 1),
(1, 'Lower Body Focus', 'Focus on legs, glutes and core', 3),
(1, 'Full Body Power', 'Compound movements for overall strength', 5)
ON CONFLICT DO NOTHING;

-- Insert sample exercises
INSERT INTO exercises (workout_id, name, weight, repetitions, series, rest_time, notes) VALUES
-- Upper Body Workout (workout_id = 1)
(1, 'Bench Press', 80.0, 8, 3, 120, 'Focus on controlled movement'),
(1, 'Pull-ups', 0.0, 10, 3, 90, 'Use assistance if needed'),
(1, 'Shoulder Press', 40.0, 10, 3, 90, 'Keep core engaged'),
(1, 'Barbell Rows', 60.0, 10, 3, 90, 'Squeeze shoulder blades together'),

-- Lower Body Workout (workout_id = 2)
(2, 'Squats', 100.0, 8, 4, 120, 'Go to parallel or below'),
(2, 'Deadlifts', 120.0, 6, 3, 180, 'Keep back straight'),
(2, 'Lunges', 20.0, 12, 3, 60, 'Alternate legs each set'),
(2, 'Calf Raises', 60.0, 15, 3, 45, 'Full range of motion'),

-- Full Body Workout (workout_id = 3)
(3, 'Clean and Press', 50.0, 5, 4, 150, 'Explosive movement'),
(3, 'Turkish Get-ups', 16.0, 5, 3, 120, 'Each side'),
(3, 'Burpees', 0.0, 10, 3, 90, 'Maintain good form'),
(3, 'Mountain Climbers', 0.0, 20, 3, 60, 'Keep hips level')
ON CONFLICT DO NOTHING;

-- Insert sample recipe
INSERT INTO recipes (user_id, name, description, instructions, servings, prep_time, cook_time) VALUES
('demo-user', 'Protein Power Bowl', 'High protein meal with chicken and vegetables', 
'1. Cook brown rice according to package instructions. 2. Season and grill chicken breast. 3. Steam broccoli until tender. 4. Combine all ingredients in a bowl and serve.', 
2, 15, 25)
ON CONFLICT DO NOTHING;

-- Insert sample recipe ingredients
INSERT INTO recipe_ingredients (recipe_id, food_id, quantity, unit) VALUES
(1, 1, 200, 'g'),  -- Chicken Breast
(1, 2, 150, 'g'),  -- Brown Rice
(1, 3, 100, 'g')   -- Broccoli
ON CONFLICT DO NOTHING;

-- Insert sample nutrition goals
INSERT INTO nutrition_goals (user_id, daily_calories, daily_protein, daily_carbs, daily_fat, daily_fiber) VALUES
('demo-user', 2200, 165.0, 275.0, 73.0, 28.0)
ON CONFLICT (user_id) DO NOTHING;