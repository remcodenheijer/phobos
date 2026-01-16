package middleware

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// dbHolder wraps the database to prevent Fiber from closing it
// when releasing the context (Fiber calls Close() on io.Closer values in Locals)
type dbHolder struct {
	db *sql.DB
}

// Setup configures all middleware for the application
func Setup(app *fiber.App, db *sql.DB) {
	holder := &dbHolder{db: db}

	// Recovery middleware
	app.Use(recover.New())

	// Logger middleware
	app.Use(logger.New(logger.Config{
		Format: "${time} ${method} ${path} ${status} ${latency}\n",
	}))

	// Store database holder in context (not the db directly)
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("dbHolder", holder)
		return c.Next()
	})
}

// GetDB retrieves the database from context
func GetDB(c *fiber.Ctx) *sql.DB {
	return c.Locals("dbHolder").(*dbHolder).db
}
