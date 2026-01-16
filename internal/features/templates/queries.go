package templates

import (
	"database/sql"
	"fmt"
)

// ListAll returns all workout templates
func ListAll(db *sql.DB) ([]WorkoutTemplate, error) {
	rows, err := db.Query(`
		SELECT id, name, created_at, updated_at
		FROM workout_templates
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	var templates []WorkoutTemplate
	for rows.Next() {
		var t WorkoutTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		templates = append(templates, t)
	}

	return templates, rows.Err()
}

// GetByID returns a single template with its exercises
func GetByID(db *sql.DB, id int64) (*WorkoutTemplate, error) {
	var t WorkoutTemplate
	err := db.QueryRow(`
		SELECT id, name, created_at, updated_at
		FROM workout_templates
		WHERE id = ?
	`, id).Scan(&t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Load exercises
	t.Exercises, err = GetTemplateExercises(db, id)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// Create inserts a new template and returns its ID
func Create(db *sql.DB, name string) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO workout_templates (name) VALUES (?)
	`, name)
	if err != nil {
		return 0, fmt.Errorf("failed to create template: %w", err)
	}

	return result.LastInsertId()
}

// Update modifies a template's name
func Update(db *sql.DB, id int64, name string) error {
	_, err := db.Exec(`
		UPDATE workout_templates
		SET name = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, name, id)
	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}
	return nil
}

// Delete removes a template by ID
func Delete(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM workout_templates WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}
	return nil
}

// GetTemplateExercises returns all exercises for a template
func GetTemplateExercises(db *sql.DB, templateID int64) ([]TemplateExercise, error) {
	rows, err := db.Query(`
		SELECT te.id, te.template_id, te.exercise_id, te.target_sets, te.target_reps, te.position,
		       e.id, e.name, e.created_at
		FROM template_exercises te
		JOIN exercises e ON te.exercise_id = e.id
		WHERE te.template_id = ?
		ORDER BY te.position ASC
	`, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template exercises: %w", err)
	}
	defer rows.Close()

	var templateExercises []TemplateExercise
	for rows.Next() {
		var te TemplateExercise
		if err := rows.Scan(
			&te.ID, &te.TemplateID, &te.ExerciseID, &te.TargetSets, &te.TargetReps, &te.Position,
			&te.Exercise.ID, &te.Exercise.Name, &te.Exercise.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan template exercise: %w", err)
		}
		templateExercises = append(templateExercises, te)
	}

	return templateExercises, rows.Err()
}

// AddExercise adds an exercise to a template
func AddExercise(db *sql.DB, templateID, exerciseID int64, targetSets, targetReps int) (int64, error) {
	// Get the next position
	var maxPos sql.NullInt64
	err := db.QueryRow(`
		SELECT MAX(position) FROM template_exercises WHERE template_id = ?
	`, templateID).Scan(&maxPos)
	if err != nil {
		return 0, fmt.Errorf("failed to get max position: %w", err)
	}

	nextPos := 1
	if maxPos.Valid {
		nextPos = int(maxPos.Int64) + 1
	}

	result, err := db.Exec(`
		INSERT INTO template_exercises (template_id, exercise_id, target_sets, target_reps, position)
		VALUES (?, ?, ?, ?, ?)
	`, templateID, exerciseID, targetSets, targetReps, nextPos)
	if err != nil {
		return 0, fmt.Errorf("failed to add exercise to template: %w", err)
	}

	// Update template timestamp
	db.Exec(`UPDATE workout_templates SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`, templateID)

	return result.LastInsertId()
}

// UpdateExercise updates a template exercise's targets
func UpdateExercise(db *sql.DB, id int64, targetSets, targetReps int) error {
	_, err := db.Exec(`
		UPDATE template_exercises
		SET target_sets = ?, target_reps = ?
		WHERE id = ?
	`, targetSets, targetReps, id)
	if err != nil {
		return fmt.Errorf("failed to update template exercise: %w", err)
	}
	return nil
}

// RemoveExercise removes an exercise from a template
func RemoveExercise(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM template_exercises WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to remove exercise from template: %w", err)
	}
	return nil
}

// ReorderExercises updates the positions of exercises in a template
func ReorderExercises(db *sql.DB, templateID int64, exerciseIDs []int64) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for i, id := range exerciseIDs {
		_, err := tx.Exec(`
			UPDATE template_exercises
			SET position = ?
			WHERE id = ? AND template_id = ?
		`, i+1, id, templateID)
		if err != nil {
			return fmt.Errorf("failed to update position: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetExerciseByID returns a single template exercise
func GetExerciseByID(db *sql.DB, id int64) (*TemplateExercise, error) {
	var te TemplateExercise
	err := db.QueryRow(`
		SELECT te.id, te.template_id, te.exercise_id, te.target_sets, te.target_reps, te.position,
		       e.id, e.name, e.created_at
		FROM template_exercises te
		JOIN exercises e ON te.exercise_id = e.id
		WHERE te.id = ?
	`, id).Scan(
		&te.ID, &te.TemplateID, &te.ExerciseID, &te.TargetSets, &te.TargetReps, &te.Position,
		&te.Exercise.ID, &te.Exercise.Name, &te.Exercise.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get template exercise: %w", err)
	}

	return &te, nil
}
