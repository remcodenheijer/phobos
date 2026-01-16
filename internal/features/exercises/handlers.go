package exercises

import (
	"strconv"

	"phobos/internal/shared/htmx"
	"phobos/internal/shared/middleware"

	"github.com/gofiber/fiber/v2"
)

// HandleList displays all exercises
func HandleList(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	exercises, err := ListAll(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load exercises")
	}

	return htmx.Render(c, ExercisesPage(exercises))
}

// HandleCreate creates a new exercise
func HandleCreate(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	name := c.FormValue("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Name is required")
	}

	id, err := Create(db, name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to create exercise")
	}

	exercise := Exercise{ID: id, Name: name}
	return htmx.Render(c, ExerciseRow(exercise))
}

// HandleDelete removes an exercise
func HandleDelete(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	if err := Delete(db, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete exercise")
	}

	return c.SendString("") // Return empty to remove the element
}

// HandleSearch searches exercises by name
func HandleSearch(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	query := c.Query("q")
	exercises, err := Search(db, query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to search exercises")
	}

	return htmx.Render(c, ExerciseListFragment(exercises))
}
