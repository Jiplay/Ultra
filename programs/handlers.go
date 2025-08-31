package programs

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"ultra/database"
)

func CreateProgramHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateProgramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	tx, err := database.PostgresDB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Create program
	var programID int
	err = tx.QueryRow(`
		INSERT INTO programs (user_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id
	`, userID, req.Name, req.Description).Scan(&programID)
	
	if err != nil {
		http.Error(w, "Failed to create program", http.StatusInternalServerError)
		return
	}

	program := Program{
		ID:          programID,
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Workouts:    []Workout{},
	}

	// Create workouts and exercises
	for _, workoutReq := range req.Workouts {
		var workoutID int
		err = tx.QueryRow(`
			INSERT INTO workouts (program_id, name, description, day_of_week)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, programID, workoutReq.Name, workoutReq.Description, workoutReq.DayOfWeek).Scan(&workoutID)
		
		if err != nil {
			http.Error(w, "Failed to create workout", http.StatusInternalServerError)
			return
		}

		workout := Workout{
			ID:          workoutID,
			ProgramID:   programID,
			Name:        workoutReq.Name,
			Description: workoutReq.Description,
			DayOfWeek:   workoutReq.DayOfWeek,
			Exercises:   []Exercise{},
		}

		// Create exercises
		for _, exerciseReq := range workoutReq.Exercises {
			var exerciseID int
			err = tx.QueryRow(`
				INSERT INTO exercises (workout_id, name, weight, repetitions, series, rest_time, notes)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				RETURNING id
			`, workoutID, exerciseReq.Name, exerciseReq.Weight, exerciseReq.Repetitions, 
			   exerciseReq.Series, exerciseReq.RestTime, exerciseReq.Notes).Scan(&exerciseID)
			
			if err != nil {
				http.Error(w, "Failed to create exercise", http.StatusInternalServerError)
				return
			}

			exercise := Exercise{
				ID:          exerciseID,
				WorkoutID:   workoutID,
				Name:        exerciseReq.Name,
				Weight:      exerciseReq.Weight,
				Repetitions: exerciseReq.Repetitions,
				Series:      exerciseReq.Series,
				RestTime:    exerciseReq.RestTime,
				Notes:       exerciseReq.Notes,
			}

			workout.Exercises = append(workout.Exercises, exercise)
		}

		program.Workouts = append(program.Workouts, workout)
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(program)
}

func GetProgramsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	rows, err := database.PostgresDB.Query(`
		SELECT id, user_id, name, description, created_at, updated_at
		FROM programs
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		http.Error(w, "Failed to get programs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var programs []Program
	for rows.Next() {
		var program Program
		err := rows.Scan(&program.ID, &program.UserID, &program.Name, 
						&program.Description, &program.CreatedAt, &program.UpdatedAt)
		if err != nil {
			http.Error(w, "Failed to scan program", http.StatusInternalServerError)
			return
		}
		programs = append(programs, program)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(programs)
}

func GetProgramHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	programID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid program ID", http.StatusBadRequest)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	var program Program
	err = database.PostgresDB.QueryRow(`
		SELECT id, user_id, name, description, created_at, updated_at
		FROM programs
		WHERE id = $1 AND user_id = $2
	`, programID, userID).Scan(&program.ID, &program.UserID, &program.Name,
		&program.Description, &program.CreatedAt, &program.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Program not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get program", http.StatusInternalServerError)
		return
	}

	// Get workouts
	workoutRows, err := database.PostgresDB.Query(`
		SELECT id, program_id, name, description, day_of_week, created_at
		FROM workouts
		WHERE program_id = $1
		ORDER BY day_of_week, created_at
	`, programID)
	if err != nil {
		http.Error(w, "Failed to get workouts", http.StatusInternalServerError)
		return
	}
	defer workoutRows.Close()

	program.Workouts = []Workout{}
	for workoutRows.Next() {
		var workout Workout
		err := workoutRows.Scan(&workout.ID, &workout.ProgramID, &workout.Name,
							   &workout.Description, &workout.DayOfWeek, &workout.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to scan workout", http.StatusInternalServerError)
			return
		}

		// Get exercises for this workout
		exerciseRows, err := database.PostgresDB.Query(`
			SELECT id, workout_id, name, weight, repetitions, series, rest_time, notes, created_at
			FROM exercises
			WHERE workout_id = $1
			ORDER BY created_at
		`, workout.ID)
		if err != nil {
			http.Error(w, "Failed to get exercises", http.StatusInternalServerError)
			return
		}
		defer exerciseRows.Close()

		workout.Exercises = []Exercise{}
		for exerciseRows.Next() {
			var exercise Exercise
			err := exerciseRows.Scan(&exercise.ID, &exercise.WorkoutID, &exercise.Name,
								     &exercise.Weight, &exercise.Repetitions, &exercise.Series,
								     &exercise.RestTime, &exercise.Notes, &exercise.CreatedAt)
			if err != nil {
				http.Error(w, "Failed to scan exercise", http.StatusInternalServerError)
				return
			}
			workout.Exercises = append(workout.Exercises, exercise)
		}

		program.Workouts = append(program.Workouts, workout)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(program)
}

func UpdateProgramHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	programID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid program ID", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	var req UpdateProgramRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

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
	args = append(args, programID)

	argCount++
	query += " AND user_id = $" + strconv.Itoa(argCount)
	args = append(args, userID)

	result, err := database.PostgresDB.Exec(query, args...)
	if err != nil {
		http.Error(w, "Failed to update program", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to check update result", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Program not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteProgramHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	programID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid program ID", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	result, err := database.PostgresDB.Exec(`
		DELETE FROM programs
		WHERE id = $1 AND user_id = $2
	`, programID, userID)
	
	if err != nil {
		http.Error(w, "Failed to delete program", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Failed to check delete result", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Program not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func RegisterRoutes(router *mux.Router) {
	programRouter := router.PathPrefix("/programs").Subrouter()
	
	programRouter.HandleFunc("", CreateProgramHandler).Methods("POST")
	programRouter.HandleFunc("", GetProgramsHandler).Methods("GET")
	programRouter.HandleFunc("/{id}", GetProgramHandler).Methods("GET")
	programRouter.HandleFunc("/{id}", UpdateProgramHandler).Methods("PUT")
	programRouter.HandleFunc("/{id}", DeleteProgramHandler).Methods("DELETE")
}