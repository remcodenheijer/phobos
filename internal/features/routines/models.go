package routines

import (
	"phobos/internal/features/templates"
	"time"
)

// Routine represents a collection of workout templates
type Routine struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Templates []RoutineTemplate
}

// RoutineTemplate represents a template in a routine
type RoutineTemplate struct {
	ID         int64
	RoutineID  int64
	TemplateID int64
	Template   templates.WorkoutTemplate
	Position   int
}
