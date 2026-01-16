-- +goose Up
-- exercises: Named movements
CREATE TABLE exercises (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- workout_templates: Reusable blueprints
CREATE TABLE workout_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- template_exercises: Exercises in templates with targets
CREATE TABLE template_exercises (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INTEGER NOT NULL REFERENCES workout_templates(id) ON DELETE CASCADE,
    exercise_id INTEGER NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
    target_sets INTEGER NOT NULL CHECK (target_sets > 0),
    target_reps INTEGER NOT NULL CHECK (target_reps > 0),
    position INTEGER NOT NULL,
    UNIQUE(template_id, position)
);

-- workouts: Training sessions
CREATE TABLE workouts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    date DATE NOT NULL,
    notes TEXT,
    status TEXT NOT NULL CHECK (status IN ('in_progress', 'finished')) DEFAULT 'in_progress',
    template_id INTEGER REFERENCES workout_templates(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMP
);

-- workout_exercises: Exercises in a workout
CREATE TABLE workout_exercises (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workout_id INTEGER NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    exercise_id INTEGER NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
    position INTEGER NOT NULL,
    UNIQUE(workout_id, position)
);

-- logged_sets: Individual sets performed
CREATE TABLE logged_sets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workout_exercise_id INTEGER NOT NULL REFERENCES workout_exercises(id) ON DELETE CASCADE,
    reps INTEGER NOT NULL CHECK (reps >= 0),
    weight REAL NOT NULL CHECK (weight >= 0),
    position INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(workout_exercise_id, position)
);

-- routines: Collections of templates
CREATE TABLE routines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- routine_templates: Templates in routines
CREATE TABLE routine_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    routine_id INTEGER NOT NULL REFERENCES routines(id) ON DELETE CASCADE,
    template_id INTEGER NOT NULL REFERENCES workout_templates(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    UNIQUE(routine_id, position)
);

-- +goose Down
DROP TABLE IF EXISTS routine_templates;
DROP TABLE IF EXISTS routines;
DROP TABLE IF EXISTS logged_sets;
DROP TABLE IF EXISTS workout_exercises;
DROP TABLE IF EXISTS workouts;
DROP TABLE IF EXISTS template_exercises;
DROP TABLE IF EXISTS workout_templates;
DROP TABLE IF EXISTS exercises;
