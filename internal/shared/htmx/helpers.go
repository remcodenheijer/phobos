package htmx

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

// IsHTMX checks if the request is an HTMX request
func IsHTMX(c *fiber.Ctx) bool {
	return c.Get("HX-Request") == "true"
}

// Render renders a templ component to the response
func Render(c *fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	handler := adaptor.HTTPHandler(templ.Handler(component))
	return handler(c)
}

// Redirect sends an HX-Redirect header for HTMX requests, or a standard redirect otherwise
func Redirect(c *fiber.Ctx, path string) error {
	if IsHTMX(c) {
		c.Set("HX-Redirect", path)
		return c.SendStatus(fiber.StatusOK)
	}
	return c.Redirect(path, fiber.StatusSeeOther)
}

// Refresh triggers a page refresh via HX-Refresh header
func Refresh(c *fiber.Ctx) error {
	c.Set("HX-Refresh", "true")
	return c.SendStatus(fiber.StatusOK)
}

// Trigger sends an HX-Trigger header to trigger client-side events
func Trigger(c *fiber.Ctx, event string) {
	c.Set("HX-Trigger", event)
}

// Retarget changes the target element for the response
func Retarget(c *fiber.Ctx, selector string) {
	c.Set("HX-Retarget", selector)
}

// Reswap changes the swap behavior for the response
func Reswap(c *fiber.Ctx, method string) {
	c.Set("HX-Reswap", method)
}
