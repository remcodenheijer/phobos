package workouts

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ListInProgress returns all in-progress workouts
func ListInProgress(db *sql.DB) ([]WorkoutSummary, error) {
	return listByStatus(db, StatusInProgress)
}

// ListFinished returns all finished workouts (newest first)
func ListFinished(db *sql.DB) ([]WorkoutSummary, error) {
	return listByStatus(db, StatusFinished)
}

func listByStatus(db *sql.DB, status WorkoutStatus) ([]WorkoutSummary, error) {
	rows, err := db.Query(`
		SELECT w.id, w.name, w.date, w.status,
		       COUNT(DISTINCT we.id) as exercise_count,
		       COUNT(ls.id) as set_count
		FROM workouts w
		LEFT JOIN workout_exercises we ON we.workout_id = w.id
		LEFT JOIN logged_sets ls ON ls.workout_exercise_id = we.id
		WHERE w.status = ?
		GROUP BY w.id, w.name, w.date, w.status, w.created_at
		ORDER BY w.date DESC, w.created_at DESC
	`, status)
	if err != nil {
		return nil, fmt.Errorf("failed to list workouts: %w", err)
	}
	defer rows.Close()

	var workouts []WorkoutSummary
	for rows.Next() {
		var w WorkoutSummary
		if err := rows.Scan(&w.ID, &w.Name, &w.Date, &w.Status, &w.ExerciseCount, &w.SetCount); err != nil {
			return nil, fmt.Errorf("failed to scan workout: %w", err)
		}
		workouts = append(workouts, w)
	}

	return workouts, rows.Err()
}

// GetByID returns a single workout with all exercises and sets
func GetByID(db *sql.DB, id int64) (*Workout, error) {
	var w Workout
	var notes sql.NullString
	var templateID sql.NullInt64
	var finishedAt sql.NullTime

	err := db.QueryRow(`
		SELECT id, name, date, notes, status, template_id, created_at, finished_at
		FROM workouts
		WHERE id = ?
	`, id).Scan(&w.ID, &w.Name, &w.Date, &notes, &w.Status, &templateID, &w.CreatedAt, &finishedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workout: %w", err)
	}

	w.Notes = notes.String
	if templateID.Valid {
		w.TemplateID = &templateID.Int64
	}
	if finishedAt.Valid {
		w.FinishedAt = &finishedAt.Time
	}

	// Load exercises
	w.Exercises, err = GetWorkoutExercises(db, id)
	if err != nil {
		return nil, err
	}

	return &w, nil
}

// Create inserts a new workout and returns its ID
func Create(db *sql.DB, name string, date time.Time, templateID *int64) (int64, error) {
	var result sql.Result
	var err error

	if templateID != nil {
		result, err = db.Exec(`
			INSERT INTO workouts (name, date, template_id, status)
			VALUES (?, ?, ?, ?)
		`, name, date.Format("2006-01-02"), *templateID, StatusInProgress)
	} else {
		result, err = db.Exec(`
			INSERT INTO workouts (name, date, status)
			VALUES (?, ?, ?)
		`, name, date.Format("2006-01-02"), StatusInProgress)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to create workout: %w", err)
	}

	return result.LastInsertId()
}

// Update modifies a workout's details
func Update(db *sql.DB, id int64, name string, date time.Time, notes string) error {
	_, err := db.Exec(`
		UPDATE workouts
		SET name = ?, date = ?, notes = ?
		WHERE id = ? AND status = ?
	`, name, date.Format("2006-01-02"), notes, id, StatusInProgress)
	if err != nil {
		return fmt.Errorf("failed to update workout: %w", err)
	}
	return nil
}

// Delete removes a workout by ID
func Delete(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM workouts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete workout: %w", err)
	}
	return nil
}

// Finish marks a workout as complete
func Finish(db *sql.DB, id int64) error {
	_, err := db.Exec(`
		UPDATE workouts
		SET status = ?, finished_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = ?
	`, StatusFinished, id, StatusInProgress)
	if err != nil {
		return fmt.Errorf("failed to finish workout: %w", err)
	}
	return nil
}

// GetWorkoutExercises returns all exercises for a workout with sets and last weight
func GetWorkoutExercises(db *sql.DB, workoutID int64) ([]WorkoutExercise, error) {
	rows, err := db.Query(`
		SELECT we.id, we.workout_id, we.exercise_id, we.position,
		       e.id, e.name, e.created_at
		FROM workout_exercises we
		JOIN exercises e ON we.exercise_id = e.id
		WHERE we.workout_id = ?
		ORDER BY we.position ASC
	`, workoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workout exercises: %w", err)
	}
	defer rows.Close()

	var exercises []WorkoutExercise
	var workoutExerciseIDs []int64
	var exerciseIDs []int64
	for rows.Next() {
		var we WorkoutExercise
		if err := rows.Scan(
			&we.ID, &we.WorkoutID, &we.ExerciseID, &we.Position,
			&we.Exercise.ID, &we.Exercise.Name, &we.Exercise.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan workout exercise: %w", err)
		}
		exercises = append(exercises, we)
		workoutExerciseIDs = append(workoutExerciseIDs, we.ID)
		exerciseIDs = append(exerciseIDs, we.ExerciseID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(exercises) == 0 {
		return exercises, nil
	}

	// Batch load all sets for all workout exercises
	setsMap, err := getLoggedSetsBatch(db, workoutExerciseIDs)
	if err != nil {
		return nil, err
	}

	// Batch load last weights for all exercises
	lastWeightsMap, err := getLastWeightsBatch(db, exerciseIDs, workoutID)
	if err != nil {
		return nil, err
	}

	// Assign sets and last weights to exercises
	for i := range exercises {
		exercises[i].Sets = setsMap[exercises[i].ID]
		exercises[i].LastWeight = lastWeightsMap[exercises[i].ExerciseID]
	}

	return exercises, nil
}

// getLoggedSetsBatch fetches all logged sets for multiple workout exercises in one query
func getLoggedSetsBatch(db *sql.DB, workoutExerciseIDs []int64) (map[int64][]LoggedSet, error) {
	if len(workoutExerciseIDs) == 0 {
		return make(map[int64][]LoggedSet), nil
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(workoutExerciseIDs))
	args := make([]interface{}, len(workoutExerciseIDs))
	for i, id := range workoutExerciseIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, workout_exercise_id, reps, weight, position, created_at
		FROM logged_sets
		WHERE workout_exercise_id IN (%s)
		ORDER BY workout_exercise_id, position ASC
	`, strings.Join(placeholders, ","))

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get logged sets: %w", err)
	}
	defer rows.Close()

	result := make(map[int64][]LoggedSet)
	for rows.Next() {
		var s LoggedSet
		if err := rows.Scan(&s.ID, &s.WorkoutExerciseID, &s.Reps, &s.Weight, &s.Position, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan logged set: %w", err)
		}
		result[s.WorkoutExerciseID] = append(result[s.WorkoutExerciseID], s)
	}

	return result, rows.Err()
}

// getLastWeightsBatch fetches the most recent weight for multiple exercises in one query
func getLastWeightsBatch(db *sql.DB, exerciseIDs []int64, excludeWorkoutID int64) (map[int64]*float64, error) {
	if len(exerciseIDs) == 0 {
		return make(map[int64]*float64), nil
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(exerciseIDs))
	args := make([]interface{}, len(exerciseIDs))
	for i, id := range exerciseIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	args = append(args, StatusFinished, excludeWorkoutID)

	// Use a subquery to get the most recent weight per exercise
	query := fmt.Sprintf(`
		SELECT we.exercise_id, ls.weight
		FROM logged_sets ls
		JOIN workout_exercises we ON ls.workout_exercise_id = we.id
		JOIN workouts w ON we.workout_id = w.id
		WHERE we.exercise_id IN (%s)
		  AND w.status = ?
		  AND w.id != ?
		  AND ls.id = (
		    SELECT ls2.id
		    FROM logged_sets ls2
		    JOIN workout_exercises we2 ON ls2.workout_exercise_id = we2.id
		    JOIN workouts w2 ON we2.workout_id = w2.id
		    WHERE we2.exercise_id = we.exercise_id
		      AND w2.status = ?
		      AND w2.id != ?
		    ORDER BY w2.date DESC, ls2.created_at DESC
		    LIMIT 1
		  )
	`, strings.Join(placeholders, ","))

	args = append(args, StatusFinished, excludeWorkoutID)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get last weights: %w", err)
	}
	defer rows.Close()

	result := make(map[int64]*float64)
	for rows.Next() {
		var exerciseID int64
		var weight float64
		if err := rows.Scan(&exerciseID, &weight); err != nil {
			return nil, fmt.Errorf("failed to scan last weight: %w", err)
		}
		w := weight
		result[exerciseID] = &w
	}

	return result, rows.Err()
}

// AddExercise adds an exercise to a workout
func AddExercise(db *sql.DB, workoutID, exerciseID int64) (int64, error) {
	// Get the next position
	var maxPos sql.NullInt64
	err := db.QueryRow(`
		SELECT MAX(position) FROM workout_exercises WHERE workout_id = ?
	`, workoutID).Scan(&maxPos)
	if err != nil {
		return 0, fmt.Errorf("failed to get max position: %w", err)
	}

	nextPos := 1
	if maxPos.Valid {
		nextPos = int(maxPos.Int64) + 1
	}

	result, err := db.Exec(`
		INSERT INTO workout_exercises (workout_id, exercise_id, position)
		VALUES (?, ?, ?)
	`, workoutID, exerciseID, nextPos)
	if err != nil {
		return 0, fmt.Errorf("failed to add exercise to workout: %w", err)
	}

	return result.LastInsertId()
}

// RemoveExercise removes an exercise from a workout
func RemoveExercise(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM workout_exercises WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to remove exercise from workout: %w", err)
	}
	return nil
}

// GetLoggedSets returns all sets for a workout exercise
func GetLoggedSets(db *sql.DB, workoutExerciseID int64) ([]LoggedSet, error) {
	rows, err := db.Query(`
		SELECT id, workout_exercise_id, reps, weight, position, created_at
		FROM logged_sets
		WHERE workout_exercise_id = ?
		ORDER BY position ASC
	`, workoutExerciseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get logged sets: %w", err)
	}
	defer rows.Close()

	var sets []LoggedSet
	for rows.Next() {
		var s LoggedSet
		if err := rows.Scan(&s.ID, &s.WorkoutExerciseID, &s.Reps, &s.Weight, &s.Position, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan logged set: %w", err)
		}
		sets = append(sets, s)
	}

	return sets, rows.Err()
}

// AddSet adds a new set to a workout exercise
func AddSet(db *sql.DB, workoutExerciseID int64, reps int, weight float64) (int64, error) {
	// Get the next position
	var maxPos sql.NullInt64
	err := db.QueryRow(`
		SELECT MAX(position) FROM logged_sets WHERE workout_exercise_id = ?
	`, workoutExerciseID).Scan(&maxPos)
	if err != nil {
		return 0, fmt.Errorf("failed to get max position: %w", err)
	}

	nextPos := 1
	if maxPos.Valid {
		nextPos = int(maxPos.Int64) + 1
	}

	result, err := db.Exec(`
		INSERT INTO logged_sets (workout_exercise_id, reps, weight, position)
		VALUES (?, ?, ?, ?)
	`, workoutExerciseID, reps, weight, nextPos)
	if err != nil {
		return 0, fmt.Errorf("failed to add set: %w", err)
	}

	return result.LastInsertId()
}

// UpdateSet modifies an existing set
func UpdateSet(db *sql.DB, id int64, reps int, weight float64) error {
	_, err := db.Exec(`
		UPDATE logged_sets
		SET reps = ?, weight = ?
		WHERE id = ?
	`, reps, weight, id)
	if err != nil {
		return fmt.Errorf("failed to update set: %w", err)
	}
	return nil
}

// DeleteSet removes a set
func DeleteSet(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM logged_sets WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete set: %w", err)
	}
	return nil
}

// GetSetByID returns a single set
func GetSetByID(db *sql.DB, id int64) (*LoggedSet, error) {
	var s LoggedSet
	err := db.QueryRow(`
		SELECT id, workout_exercise_id, reps, weight, position, created_at
		FROM logged_sets
		WHERE id = ?
	`, id).Scan(&s.ID, &s.WorkoutExerciseID, &s.Reps, &s.Weight, &s.Position, &s.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get set: %w", err)
	}

	return &s, nil
}

// GetLastWeight returns the most recent weight used for an exercise from finished workouts
func GetLastWeight(db *sql.DB, exerciseID, excludeWorkoutID int64) (*float64, error) {
	var weight sql.NullFloat64
	err := db.QueryRow(`
		SELECT ls.weight
		FROM logged_sets ls
		JOIN workout_exercises we ON ls.workout_exercise_id = we.id
		JOIN workouts w ON we.workout_id = w.id
		WHERE we.exercise_id = ?
		  AND w.status = ?
		  AND w.id != ?
		ORDER BY w.date DESC, ls.created_at DESC
		LIMIT 1
	`, exerciseID, StatusFinished, excludeWorkoutID).Scan(&weight)

	if err == sql.ErrNoRows || !weight.Valid {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get last weight: %w", err)
	}

	return &weight.Float64, nil
}

// GetWorkoutExerciseByID returns a single workout exercise
func GetWorkoutExerciseByID(db *sql.DB, id int64) (*WorkoutExercise, error) {
	var we WorkoutExercise
	err := db.QueryRow(`
		SELECT we.id, we.workout_id, we.exercise_id, we.position,
		       e.id, e.name, e.created_at
		FROM workout_exercises we
		JOIN exercises e ON we.exercise_id = e.id
		WHERE we.id = ?
	`, id).Scan(
		&we.ID, &we.WorkoutID, &we.ExerciseID, &we.Position,
		&we.Exercise.ID, &we.Exercise.Name, &we.Exercise.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workout exercise: %w", err)
	}

	sets, err := GetLoggedSets(db, we.ID)
	if err != nil {
		return nil, err
	}
	we.Sets = sets

	lastWeight, err := GetLastWeight(db, we.ExerciseID, we.WorkoutID)
	if err != nil {
		return nil, err
	}
	we.LastWeight = lastWeight

	return &we, nil
}

// CreateFromTemplate creates a workout from a template
func CreateFromTemplate(db *sql.DB, workoutName string, date time.Time, templateID int64) (int64, error) {
	// Create the workout
	workoutID, err := Create(db, workoutName, date, &templateID)
	if err != nil {
		return 0, err
	}

	// Copy exercises from template - first collect all exercises to avoid holding
	// the cursor open while doing inserts (which would deadlock with MaxOpenConns=1)
	rows, err := db.Query(`
		SELECT exercise_id, target_sets, target_reps, position
		FROM template_exercises
		WHERE template_id = ?
		ORDER BY position ASC
	`, templateID)
	if err != nil {
		return 0, fmt.Errorf("failed to get template exercises: %w", err)
	}

	type templateExercise struct {
		exerciseID int64
		position   int
	}
	var exercises []templateExercise

	for rows.Next() {
		var exerciseID int64
		var targetSets, targetReps, position int
		if err := rows.Scan(&exerciseID, &targetSets, &targetReps, &position); err != nil {
			rows.Close()
			return 0, fmt.Errorf("failed to scan template exercise: %w", err)
		}
		exercises = append(exercises, templateExercise{exerciseID: exerciseID, position: position})
	}
	rows.Close()

	if err := rows.Err(); err != nil {
		return 0, err
	}

	// Now insert the exercises
	for _, ex := range exercises {
		_, err := db.Exec(`
			INSERT INTO workout_exercises (workout_id, exercise_id, position)
			VALUES (?, ?, ?)
		`, workoutID, ex.exerciseID, ex.position)
		if err != nil {
			return 0, fmt.Errorf("failed to copy exercise: %w", err)
		}
	}

	return workoutID, nil
}

// GetTemplateTargets returns the target sets/reps for an exercise from a template
func GetTemplateTargets(db *sql.DB, templateID, exerciseID int64) (*int, *int, error) {
	var targetSets, targetReps int
	err := db.QueryRow(`
		SELECT target_sets, target_reps
		FROM template_exercises
		WHERE template_id = ? AND exercise_id = ?
	`, templateID, exerciseID).Scan(&targetSets, &targetReps)

	if err == sql.ErrNoRows {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get template targets: %w", err)
	}

	return &targetSets, &targetReps, nil
}
