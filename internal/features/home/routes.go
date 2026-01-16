package home

import "github.com/gofiber/fiber/v2"

// RegisterRoutes sets up home routes
func RegisterRoutes(app *fiber.App) {
	app.Get("/", HandleHome)
}
