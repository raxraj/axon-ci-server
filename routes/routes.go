package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/raxraj/axon-ci-server/routes/users"
)

func Routes(e *echo.Echo) {
	// Static files
	e.Static("/static", "static")

	// Health check route
	e.GET("/health-check", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	// Version route
	e.GET("/version", func(c echo.Context) error {
		return c.String(200, "1.0.0")
	})

	// API groups
	api := e.Group("/v1")

	// Routes
	users.Routes(api)
}
