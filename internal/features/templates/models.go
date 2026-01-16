package templates

import (
	"phobos/internal/features/exercises"
	"time"
)

// WorkoutTemplate represents a reusable workout blueprint
type WorkoutTemplate struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Exercises []TemplateExercise
}

// TemplateExercise represents an exercise in a template with targets
type TemplateExercise struct {
	ID         int64
	TemplateID int64
	ExerciseID int64
	Exercise   exercises.Exercise
	TargetSets int
	TargetReps int
	Position   int
}
