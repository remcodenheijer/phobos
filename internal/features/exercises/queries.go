package exercises

import (
	"database/sql"
	"fmt"
)

// ListAll returns all exercises ordered by name
func ListAll(db *sql.DB) ([]Exercise, error) {
	rows, err := db.Query(`
		SELECT id, name, created_at
		FROM exercises
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list exercises: %w", err)
	}
	defer rows.Close()

	var exercises []Exercise
	for rows.Next() {
		var e Exercise
		if err := rows.Scan(&e.ID, &e.Name, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan exercise: %w", err)
		}
		exercises = append(exercises, e)
	}

	return exercises, rows.Err()
}

// GetByID returns a single exercise by ID
func GetByID(db *sql.DB, id int64) (*Exercise, error) {
	var e Exercise
	err := db.QueryRow(`
		SELECT id, name, created_at
		FROM exercises
		WHERE id = ?
	`, id).Scan(&e.ID, &e.Name, &e.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get exercise: %w", err)
	}

	return &e, nil
}

// Create inserts a new exercise and returns its ID
func Create(db *sql.DB, name string) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO exercises (name) VALUES (?)
	`, name)
	if err != nil {
		return 0, fmt.Errorf("failed to create exercise: %w", err)
	}

	return result.LastInsertId()
}

// Delete removes an exercise by ID
func Delete(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM exercises WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete exercise: %w", err)
	}
	return nil
}

// Search returns exercises matching the query
func Search(db *sql.DB, query string) ([]Exercise, error) {
	rows, err := db.Query(`
		SELECT id, name, created_at
		FROM exercises
		WHERE name LIKE ?
		ORDER BY name ASC
	`, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search exercises: %w", err)
	}
	defer rows.Close()

	var exercises []Exercise
	for rows.Next() {
		var e Exercise
		if err := rows.Scan(&e.ID, &e.Name, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan exercise: %w", err)
		}
		exercises = append(exercises, e)
	}

	return exercises, rows.Err()
}
