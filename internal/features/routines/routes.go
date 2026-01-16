package routines

import "github.com/gofiber/fiber/v2"

// RegisterRoutes sets up routine routes
func RegisterRoutes(app *fiber.App) {
	app.Get("/routines", HandleList)
	app.Post("/routines", HandleCreate)
	app.Get("/routines/:id", HandleShow)
	app.Put("/routines/:id", HandleUpdate)
	app.Delete("/routines/:id", HandleDelete)
	app.Post("/routines/:id/templates", HandleAddTemplate)
	app.Delete("/routines/templates/:id", HandleRemoveTemplate)
}
