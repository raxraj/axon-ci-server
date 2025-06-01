package users

import (
	"github.com/labstack/echo/v4"
	"github.com/raxraj/axon-ci-server/controllers/users"
)

func Routes(e *echo.Group) {
	// API Group
	var Router = e.Group("/users")

	// User Routes
	Router.GET("/OAuthURL", users.OAuthInitiate)
}
