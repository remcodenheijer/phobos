package testutil

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"phobos/internal/features/exercises"
	"phobos/internal/features/home"
	"phobos/internal/features/routines"
	"phobos/internal/features/templates"
	"phobos/internal/features/workouts"
	"phobos/internal/shared/middleware"

	"github.com/gofiber/fiber/v2"
	_ "modernc.org/sqlite"
)

// TestApp holds a test application with an in-memory database
type TestApp struct {
	App *fiber.App
	DB  *sql.DB
}

// NewTestApp creates a new test application with an in-memory database
func NewTestApp(t *testing.T) *TestApp {
	t.Helper()

	// Create in-memory database
	// Set max open connections to 1 to ensure all operations use the same connection
	// (with :memory:, each connection gets a separate database)
	db, err := sql.Open("sqlite", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	db.SetMaxOpenConns(1)

	// Run migrations inline (can't use goose with in-memory easily)
	if err := runMigrations(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).SendString(err.Error())
		},
	})

	// Setup middleware
	middleware.Setup(app, db)

	// Register all routes
	home.RegisterRoutes(app)
	exercises.RegisterRoutes(app)
	templates.RegisterRoutes(app)
	workouts.RegisterRoutes(app)
	routines.RegisterRoutes(app)

	return &TestApp{
		App: app,
		DB:  db,
	}
}

// Close cleans up the test application
func (ta *TestApp) Close() {
	ta.DB.Close()
}

// Request makes a test request to the application
func (ta *TestApp) Request(method, path string, body string) *http.Response {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if body != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, _ := ta.App.Test(req, -1)
	return resp
}

// HTMXRequest makes a test HTMX request to the application
func (ta *TestApp) HTMXRequest(method, path string, body string) *http.Response {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("HX-Request", "true")
	if body != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, _ := ta.App.Test(req, -1)
	return resp
}

// ReadBody reads the response body as a string
func ReadBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	return string(body)
}

// runMigrations runs the schema migrations on the database
func runMigrations(db *sql.DB) error {
	// Execute each CREATE TABLE statement separately
	// SQLite's db.Exec may only execute the first statement when given multiple
	statements := []string{
		`CREATE TABLE exercises (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE workout_templates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE template_exercises (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			template_id INTEGER NOT NULL REFERENCES workout_templates(id) ON DELETE CASCADE,
			exercise_id INTEGER NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
			target_sets INTEGER NOT NULL CHECK (target_sets > 0),
			target_reps INTEGER NOT NULL CHECK (target_reps > 0),
			position INTEGER NOT NULL,
			UNIQUE(template_id, position)
		)`,
		`CREATE TABLE workouts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			date DATE NOT NULL,
			notes TEXT,
			status TEXT NOT NULL CHECK (status IN ('in_progress', 'finished')) DEFAULT 'in_progress',
			template_id INTEGER REFERENCES workout_templates(id) ON DELETE SET NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			finished_at TIMESTAMP
		)`,
		`CREATE TABLE workout_exercises (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workout_id INTEGER NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
			exercise_id INTEGER NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
			position INTEGER NOT NULL,
			UNIQUE(workout_id, position)
		)`,
		`CREATE TABLE logged_sets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workout_exercise_id INTEGER NOT NULL REFERENCES workout_exercises(id) ON DELETE CASCADE,
			reps INTEGER NOT NULL CHECK (reps >= 0),
			weight REAL NOT NULL CHECK (weight >= 0),
			position INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(workout_exercise_id, position)
		)`,
		`CREATE TABLE routines (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE routine_templates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			routine_id INTEGER NOT NULL REFERENCES routines(id) ON DELETE CASCADE,
			template_id INTEGER NOT NULL REFERENCES workout_templates(id) ON DELETE CASCADE,
			position INTEGER NOT NULL,
			UNIQUE(routine_id, position)
		)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
