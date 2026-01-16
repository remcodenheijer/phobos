package workouts

import (
	"phobos/internal/features/exercises"
	"time"
)

// WorkoutStatus represents the status of a workout
type WorkoutStatus string

const (
	StatusInProgress WorkoutStatus = "in_progress"
	StatusFinished   WorkoutStatus = "finished"
)

// Workout represents a training session
type Workout struct {
	ID         int64
	Name       string
	Date       time.Time
	Notes      string
	Status     WorkoutStatus
	TemplateID *int64
	CreatedAt  time.Time
	FinishedAt *time.Time
	Exercises  []WorkoutExercise
}

// IsFinished returns true if the workout is finished
func (w *Workout) IsFinished() bool {
	return w.Status == StatusFinished
}

// WorkoutExercise represents an exercise in a workout
type WorkoutExercise struct {
	ID         int64
	WorkoutID  int64
	ExerciseID int64
	Exercise   exercises.Exercise
	Position   int
	Sets       []LoggedSet
	LastWeight *float64 // Most recent weight used for this exercise
	TargetSets *int     // From template, if applicable
	TargetReps *int     // From template, if applicable
}

// LoggedSet represents an individual set performed
type LoggedSet struct {
	ID                int64
	WorkoutExerciseID int64
	Reps              int
	Weight            float64
	Position          int
	CreatedAt         time.Time
}

// WorkoutSummary is a condensed view for listing workouts
type WorkoutSummary struct {
	ID            int64
	Name          string
	Date          time.Time
	Status        WorkoutStatus
	ExerciseCount int
	SetCount      int
}
