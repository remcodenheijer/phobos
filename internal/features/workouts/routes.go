package workouts

import "github.com/gofiber/fiber/v2"

// RegisterRoutes sets up workout routes
func RegisterRoutes(app *fiber.App) {
	// Workout CRUD
	app.Get("/workouts", HandleList)
	app.Get("/workouts/new", HandleNew)
	app.Post("/workouts", HandleCreate)
	app.Get("/workouts/history", HandleHistory)
	app.Get("/workouts/:id", HandleShow)
	app.Put("/workouts/:id", HandleUpdate)
	app.Delete("/workouts/:id", HandleDelete)
	app.Post("/workouts/:id/finish", HandleFinish)

	// Workout exercises
	app.Post("/workouts/:id/exercises", HandleAddExercise)
	app.Delete("/workouts/exercises/:id", HandleRemoveExercise)

	// Logged sets
	app.Post("/workouts/exercises/:id/sets", HandleAddSet)
	app.Put("/workouts/sets/:id", HandleUpdateSet)
	app.Delete("/workouts/sets/:id", HandleDeleteSet)
}
