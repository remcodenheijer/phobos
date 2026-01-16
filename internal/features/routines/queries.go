package routines

import (
	"database/sql"
	"fmt"
)

// ListAll returns all routines
func ListAll(db *sql.DB) ([]Routine, error) {
	rows, err := db.Query(`
		SELECT id, name, created_at, updated_at
		FROM routines
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list routines: %w", err)
	}
	defer rows.Close()

	var routines []Routine
	for rows.Next() {
		var r Routine
		if err := rows.Scan(&r.ID, &r.Name, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan routine: %w", err)
		}
		routines = append(routines, r)
	}

	return routines, rows.Err()
}

// GetByID returns a single routine with its templates
func GetByID(db *sql.DB, id int64) (*Routine, error) {
	var r Routine
	err := db.QueryRow(`
		SELECT id, name, created_at, updated_at
		FROM routines
		WHERE id = ?
	`, id).Scan(&r.ID, &r.Name, &r.CreatedAt, &r.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get routine: %w", err)
	}

	// Load templates
	r.Templates, err = GetRoutineTemplates(db, id)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// Create inserts a new routine and returns its ID
func Create(db *sql.DB, name string) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO routines (name) VALUES (?)
	`, name)
	if err != nil {
		return 0, fmt.Errorf("failed to create routine: %w", err)
	}

	return result.LastInsertId()
}

// Update modifies a routine's name
func Update(db *sql.DB, id int64, name string) error {
	_, err := db.Exec(`
		UPDATE routines
		SET name = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, name, id)
	if err != nil {
		return fmt.Errorf("failed to update routine: %w", err)
	}
	return nil
}

// Delete removes a routine by ID
func Delete(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM routines WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete routine: %w", err)
	}
	return nil
}

// GetRoutineTemplates returns all templates for a routine
func GetRoutineTemplates(db *sql.DB, routineID int64) ([]RoutineTemplate, error) {
	rows, err := db.Query(`
		SELECT rt.id, rt.routine_id, rt.template_id, rt.position,
		       wt.id, wt.name, wt.created_at, wt.updated_at
		FROM routine_templates rt
		JOIN workout_templates wt ON rt.template_id = wt.id
		WHERE rt.routine_id = ?
		ORDER BY rt.position ASC
	`, routineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get routine templates: %w", err)
	}
	defer rows.Close()

	var routineTemplates []RoutineTemplate
	for rows.Next() {
		var rt RoutineTemplate
		if err := rows.Scan(
			&rt.ID, &rt.RoutineID, &rt.TemplateID, &rt.Position,
			&rt.Template.ID, &rt.Template.Name, &rt.Template.CreatedAt, &rt.Template.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan routine template: %w", err)
		}
		routineTemplates = append(routineTemplates, rt)
	}

	return routineTemplates, rows.Err()
}

// AddTemplate adds a template to a routine
func AddTemplate(db *sql.DB, routineID, templateID int64) (int64, error) {
	// Get the next position
	var maxPos sql.NullInt64
	err := db.QueryRow(`
		SELECT MAX(position) FROM routine_templates WHERE routine_id = ?
	`, routineID).Scan(&maxPos)
	if err != nil {
		return 0, fmt.Errorf("failed to get max position: %w", err)
	}

	nextPos := 1
	if maxPos.Valid {
		nextPos = int(maxPos.Int64) + 1
	}

	result, err := db.Exec(`
		INSERT INTO routine_templates (routine_id, template_id, position)
		VALUES (?, ?, ?)
	`, routineID, templateID, nextPos)
	if err != nil {
		return 0, fmt.Errorf("failed to add template to routine: %w", err)
	}

	// Update routine timestamp
	db.Exec(`UPDATE routines SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`, routineID)

	return result.LastInsertId()
}

// RemoveTemplate removes a template from a routine
func RemoveTemplate(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM routine_templates WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to remove template from routine: %w", err)
	}
	return nil
}

// GetRoutineTemplateByID returns a single routine template
func GetRoutineTemplateByID(db *sql.DB, id int64) (*RoutineTemplate, error) {
	var rt RoutineTemplate
	err := db.QueryRow(`
		SELECT rt.id, rt.routine_id, rt.template_id, rt.position,
		       wt.id, wt.name, wt.created_at, wt.updated_at
		FROM routine_templates rt
		JOIN workout_templates wt ON rt.template_id = wt.id
		WHERE rt.id = ?
	`, id).Scan(
		&rt.ID, &rt.RoutineID, &rt.TemplateID, &rt.Position,
		&rt.Template.ID, &rt.Template.Name, &rt.Template.CreatedAt, &rt.Template.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get routine template: %w", err)
	}

	return &rt, nil
}

// ReorderTemplates updates the positions of templates in a routine
func ReorderTemplates(db *sql.DB, routineID int64, templateIDs []int64) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for i, id := range templateIDs {
		_, err := tx.Exec(`
			UPDATE routine_templates
			SET position = ?
			WHERE id = ? AND routine_id = ?
		`, i+1, id, routineID)
		if err != nil {
			return fmt.Errorf("failed to update position: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
