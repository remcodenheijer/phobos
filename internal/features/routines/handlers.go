package routines

import (
	"strconv"

	"phobos/internal/features/templates"
	"phobos/internal/shared/htmx"
	"phobos/internal/shared/middleware"

	"github.com/gofiber/fiber/v2"
)

// HandleList displays all routines
func HandleList(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	routines, err := ListAll(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load routines")
	}

	return htmx.Render(c, RoutinesPage(routines))
}

// HandleCreate creates a new routine
func HandleCreate(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	name := c.FormValue("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Name is required")
	}

	id, err := Create(db, name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to create routine")
	}

	routine := Routine{ID: id, Name: name}
	return htmx.Render(c, RoutineCard(routine))
}

// HandleShow displays a single routine
func HandleShow(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	routine, err := GetByID(db, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load routine")
	}
	if routine == nil {
		return c.Status(fiber.StatusNotFound).SendString("Routine not found")
	}

	allTemplates, err := templates.ListAll(db)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load templates")
	}

	return htmx.Render(c, RoutineDetailPage(routine, allTemplates))
}

// HandleUpdate modifies a routine
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
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to update routine")
	}

	htmx.Trigger(c, "routineUpdated")
	return c.SendString("")
}

// HandleDelete removes a routine
func HandleDelete(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	if err := Delete(db, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete routine")
	}

	return c.SendString("")
}

// HandleAddTemplate adds a template to a routine
func HandleAddTemplate(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	routineID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid routine ID")
	}

	templateID, err := strconv.ParseInt(c.FormValue("template_id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid template ID")
	}

	id, err := AddTemplate(db, routineID, templateID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to add template")
	}

	rt, err := GetRoutineTemplateByID(db, id)
	if err != nil || rt == nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to load template")
	}

	return htmx.Render(c, RoutineTemplateRow(*rt))
}

// HandleRemoveTemplate removes a template from a routine
func HandleRemoveTemplate(c *fiber.Ctx) error {
	db := middleware.GetDB(c)

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	if err := RemoveTemplate(db, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to remove template")
	}

	return c.SendString("")
}
