package programs

import (
	"database/sql"
	"strconv"
)

type Repository interface {
	CreateProgram(program *Program) error
	CreateWorkout(workout *Workout) error
	CreateExercise(exercise *Exercise) error
	GetProgramsByUserID(userID string) ([]Program, error)
	GetProgramByID(id int, userID string) (*Program, error)
	GetWorkoutsByProgramID(programID int) ([]Workout, error)
	GetExercisesByWorkoutID(workoutID int) ([]Exercise, error)
	UpdateProgram(id int, userID string, req *UpdateProgramRequest) error
	DeleteProgram(id int, userID string) error
	CheckProgramExists(id int, userID string) (bool, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateProgram(program *Program) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create program
	err = tx.QueryRow(`
		INSERT INTO programs (user_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id
	`, program.UserID, program.Name, program.Description).Scan(&program.ID)
	
	if err != nil {
		return err
	}

	// Create workouts and exercises
	for i := range program.Workouts {
		workout := &program.Workouts[i]
		workout.ProgramID = program.ID
		
		err = tx.QueryRow(`
			INSERT INTO workouts (program_id, name, description, day_of_week)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, workout.ProgramID, workout.Name, workout.Description, workout.DayOfWeek).Scan(&workout.ID)
		
		if err != nil {
			return err
		}

		// Create exercises for this workout
		for j := range workout.Exercises {
			exercise := &workout.Exercises[j]
			exercise.WorkoutID = workout.ID
			
			err = tx.QueryRow(`
				INSERT INTO exercises (workout_id, name, weight, repetitions, series, rest_time, notes)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id
			`, exercise.WorkoutID, exercise.Name, exercise.Weight, exercise.Repetitions, 
			   exercise.Series, exercise.RestTime, exercise.Notes).Scan(&exercise.ID)
			
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) CreateWorkout(workout *Workout) error {
	err := r.db.QueryRow(`
		INSERT INTO workouts (program_id, name, description, day_of_week)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, workout.ProgramID, workout.Name, workout.Description, workout.DayOfWeek).Scan(&workout.ID)
	return err
}

func (r *PostgresRepository) CreateExercise(exercise *Exercise) error {
	err := r.db.QueryRow(`
		INSERT INTO exercises (workout_id, name, weight, repetitions, series, rest_time, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, exercise.WorkoutID, exercise.Name, exercise.Weight, exercise.Repetitions, 
	   exercise.Series, exercise.RestTime, exercise.Notes).Scan(&exercise.ID)
	return err
}

func (r *PostgresRepository) GetProgramsByUserID(userID string) ([]Program, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, name, description, created_at, updated_at
		FROM programs
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var programs []Program
	for rows.Next() {
		var program Program
		err := rows.Scan(&program.ID, &program.UserID, &program.Name, 
						&program.Description, &program.CreatedAt, &program.UpdatedAt)
		if err != nil {
			return nil, err
		}
		programs = append(programs, program)
	}
	return programs, nil
}

func (r *PostgresRepository) GetProgramByID(id int, userID string) (*Program, error) {
	var program Program
	err := r.db.QueryRow(`
		SELECT id, user_id, name, description, created_at, updated_at
		FROM programs
		WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(&program.ID, &program.UserID, &program.Name,
		&program.Description, &program.CreatedAt, &program.UpdatedAt)
	
	if err != nil {
		return nil, err
	}

	// Get workouts for this program
	workouts, err := r.GetWorkoutsByProgramID(id)
	if err != nil {
		return nil, err
	}
	program.Workouts = workouts

	return &program, nil
}

func (r *PostgresRepository) GetWorkoutsByProgramID(programID int) ([]Workout, error) {
	rows, err := r.db.Query(`
		SELECT id, program_id, name, description, day_of_week, created_at
		FROM workouts
		WHERE program_id = $1
		ORDER BY day_of_week, created_at
	`, programID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workouts []Workout
	for rows.Next() {
		var workout Workout
		err := rows.Scan(&workout.ID, &workout.ProgramID, &workout.Name,
						&workout.Description, &workout.DayOfWeek, &workout.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Get exercises for this workout
		exercises, err := r.GetExercisesByWorkoutID(workout.ID)
		if err != nil {
			return nil, err
		}
		workout.Exercises = exercises

		workouts = append(workouts, workout)
	}
	return workouts, nil
}

func (r *PostgresRepository) GetExercisesByWorkoutID(workoutID int) ([]Exercise, error) {
	rows, err := r.db.Query(`
		SELECT id, workout_id, name, weight, repetitions, series, rest_time, notes, created_at
		FROM exercises
		WHERE workout_id = $1
		ORDER BY created_at
	`, workoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []Exercise
	for rows.Next() {
		var exercise Exercise
		err := rows.Scan(&exercise.ID, &exercise.WorkoutID, &exercise.Name,
						&exercise.Weight, &exercise.Repetitions, &exercise.Series,
						&exercise.RestTime, &exercise.Notes, &exercise.CreatedAt)
		if err != nil {
			return nil, err
		}
		exercises = append(exercises, exercise)
	}
	return exercises, nil
}

func (r *PostgresRepository) UpdateProgram(id int, userID string, req *UpdateProgramRequest) error {
	query := "UPDATE programs SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}
	argCount := 0

	if req.Name != nil {
		argCount++
		query += ", name = $" + strconv.Itoa(argCount)
		args = append(args, *req.Name)
	}

	if req.Description != nil {
		argCount++
		query += ", description = $" + strconv.Itoa(argCount)
		args = append(args, *req.Description)
	}

	argCount++
	query += " WHERE id = $" + strconv.Itoa(argCount)
	args = append(args, id)

	argCount++
	query += " AND user_id = $" + strconv.Itoa(argCount)
	args = append(args, userID)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresRepository) DeleteProgram(id int, userID string) error {
	result, err := r.db.Exec(`
		DELETE FROM programs
		WHERE id = $1 AND user_id = $2
	`, id, userID)
	
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresRepository) CheckProgramExists(id int, userID string) (bool, error) {
	var exists int
	err := r.db.QueryRow("SELECT 1 FROM programs WHERE id = $1 AND user_id = $2", id, userID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}