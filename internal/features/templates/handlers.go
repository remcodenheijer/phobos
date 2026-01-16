package templates

import (
	"strconv"

	"phobos/internal/features/exercises"
	"phobos/internal/shared/htmx"
	"phobos/internal/shared/middleware"

	"github.com/gofiber/fiber/v2"
)

// HandleList displays all templates
func HandleList(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	templates, err := ListAll(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load templates")
	}

	return htmx.Render(c, TemplatesPage(templates))
}

// HandleCreate creates a new template
func HandleCreate(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	name := c.FormValue("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Name is required")
	}

	id, err := Create(db, name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to create template")
	}

	template := WorkoutTemplate{ID: id, Name: name}
	return htmx.Render(c, TemplateCard(template))
}

// HandleShow displays a single template
func HandleShow(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	template, err := GetByID(db, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load template")
	}
	if template == nil {
		return c.Status(fiber.StatusNotFound).SendString("Template not found")
	}

	allExercises, err := exercises.ListAll(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load exercises")
	}

	return htmx.Render(c, TemplateDetailPage(template, allExercises))
}

// HandleUpdate modifies a template
func HandleUpdate(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	name := c.FormValue("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Name is required")
	}

	if err := Update(db, id, name); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to update template")
	}

	htmx.Trigger(c, "templateUpdated")
	return c.SendString("")
}

// HandleDelete removes a template
func HandleDelete(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	if err := Delete(db, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete template")
	}

	return c.SendString("")
}

// HandleAddExercise adds an exercise to a template
func HandleAddExercise(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	templateID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid template ID")
	}

	exerciseID, err := strconv.ParseInt(c.FormValue("exercise_id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid exercise ID")
	}

	targetSets, err := strconv.Atoi(c.FormValue("target_sets"))
	if err != nil || targetSets < 1 {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid target sets")
	}

	targetReps, err := strconv.Atoi(c.FormValue("target_reps"))
	if err != nil || targetReps < 1 {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid target reps")
	}

	id, err := AddExercise(db, templateID, exerciseID, targetSets, targetReps)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to add exercise")
	}

	te, err := GetExerciseByID(db, id)
	if err != nil || te == nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load exercise")
	}

	return htmx.Render(c, TemplateExerciseRow(*te))
}

// HandleRemoveExercise removes an exercise from a template
func HandleRemoveExercise(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	if err := RemoveExercise(db, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to remove exercise")
	}

	return c.SendString("")
}
