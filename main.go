package main

import (
	"log"
	"log/slog"

	"github.com/CodedMasonry/exifmod/internal/pages"
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()
	app.Get("/:name?", func(c *fiber.Ctx) error {
		name := c.Params("name")
		c.Locals("name", name)
		if name == "" {
			name = "World"
		}
		return Render(c, pages.Home(name))
	})
	//Static File Serving
	app.Static("/assets", "./assets")

	// Middleware
	app.Use(NotFoundMiddleware)
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	slog.Info("Starting HTTP Server")
	log.Fatal(app.Listen(":8080"))
}

func NotFoundMiddleware(c *fiber.Ctx) error {
	c.Status(fiber.StatusNotFound)
	return Render(c, pages.NotFound())
}

func Render(c *fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html")
	return component.Render(c.Context(), c.Response().BodyWriter())
}
