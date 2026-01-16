package home

import (
	"phobos/internal/features/routines"
	"phobos/internal/features/templates"
	"phobos/internal/features/workouts"
	"phobos/internal/shared/htmx"
	"phobos/internal/shared/middleware"

	"github.com/gofiber/fiber/v2"
)

// DashboardData holds all data for the home page
type DashboardData struct {
	ActiveWorkouts []workouts.WorkoutSummary
	RecentWorkouts []workouts.WorkoutSummary
	Templates      []templates.WorkoutTemplate
	Routines       []routines.Routine
}

// HandleHome displays the dashboard
func HandleHome(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	// Get active workouts
	activeWorkouts, err := workouts.ListInProgress(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load active workouts")
	}

	// Get recent finished workouts (limit to 5)
	recentWorkouts, err := workouts.ListFinished(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load recent workouts")
	}
	if len(recentWorkouts) > 5 {
		recentWorkouts = recentWorkouts[:5]
	}

	// Get all templates
	allTemplates, err := templates.ListAll(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load templates")
	}

	// Get all routines
	allRoutines, err := routines.ListAll(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load routines")
	}

	data := DashboardData{
		ActiveWorkouts: activeWorkouts,
		RecentWorkouts: recentWorkouts,
		Templates:      allTemplates,
		Routines:       allRoutines,
	}

	return htmx.Render(c, HomePage(data))
}
