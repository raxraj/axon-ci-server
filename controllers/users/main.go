package users

import (
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
)

func OAuthInitiate(c echo.Context) error {
	// This function would typically initiate an OAuth flow for GitHub.
	// It returns a link to the user to GitHub's authorization page.
	params := url.Values{}
	params.Add("client_id", viper.GetString("github.client_id"))
	params.Add("redirect_uri", viper.GetString("github.redirect_uri"))
	params.Add("scope", viper.GetString("github.scope"))
	params.Add("state", viper.GetString("github.state"))
	authorizationUri := "https://github.com/login/oauth/authorize" + "?" + params.Encode()

	// Return the URL to the user in json
	return c.JSON(http.StatusOK, map[string]map[string]string{
		"data": {
			"url":     authorizationUri,
			"message": "Please visit this URL to authorize the application with GitHub.",
		},
	})
}
