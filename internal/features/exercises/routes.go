package exercises

import "github.com/gofiber/fiber/v2"

// RegisterRoutes sets up exercise routes
func RegisterRoutes(app *fiber.App) {
	app.Get("/exercises", HandleList)
	app.Post("/exercises", HandleCreate)
	app.Delete("/exercises/:id", HandleDelete)
	app.Get("/exercises/search", HandleSearch)
}
