package templates

import "github.com/gofiber/fiber/v2"

// RegisterRoutes sets up template routes
func RegisterRoutes(app *fiber.App) {
	app.Get("/templates", HandleList)
	app.Post("/templates", HandleCreate)
	app.Get("/templates/:id", HandleShow)
	app.Put("/templates/:id", HandleUpdate)
	app.Delete("/templates/:id", HandleDelete)
	app.Post("/templates/:id/exercises", HandleAddExercise)
	app.Delete("/templates/exercises/:id", HandleRemoveExercise)
}
