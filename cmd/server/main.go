package main

import (
	"flag"
	"log"
	"os"

	"phobos/internal/features/exercises"
	"phobos/internal/features/home"
	"phobos/internal/features/routines"
	"phobos/internal/features/templates"
	"phobos/internal/features/workouts"
	"phobos/internal/shared/db"
	"phobos/internal/shared/middleware"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Parse flags
	migrateOnly := flag.Bool("migrate", false, "Run migrations and exit")
	dbPath := flag.String("db", "phobos.db", "Database file path")
	port := flag.String("port", "3000", "Server port")
	flag.Parse()

	// Open database
	database, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	// Run migrations
	log.Println("Running database migrations...")
	if err := db.Migrate(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations complete")

	if *migrateOnly {
		log.Println("Migration-only mode, exiting")
		os.Exit(0)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).SendString(err.Error())
		},
	})

	// Setup middleware
	middleware.Setup(app, database)

	// Serve static files
	app.Static("/static", "./static")
	app.Static("/assets", "./assets")

	// Register routes
	home.RegisterRoutes(app)
	exercises.RegisterRoutes(app)
	templates.RegisterRoutes(app)
	workouts.RegisterRoutes(app)
	routines.RegisterRoutes(app)

	// Start server
	log.Printf("Starting server on http://localhost:%s", *port)
	if err := app.Listen(":" + *port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
