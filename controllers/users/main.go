package users

import (
	"github.com/go-resty/resty/v2"
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

type tokenResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

func OAuthCallback(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, map[string]map[string]string{
			"data": {
				"message": "missing code",
			},
		})
	}

	client := resty.New()
	result := &tokenResp{}
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetFormData(map[string]string{
			"client_id":     viper.GetString("github.client_id"),
			"client_secret": viper.GetString("github.client_secret"),
			"code":          code,
			"redirect_uri":  viper.GetString("github.redirect_uri"),
		}).
		SetResult(result).
		Post("https://github.com/login/oauth/access_token")

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]map[string]string{
			"data": {
				"message": "failed to get access token",
			},
		})
	}
	if resp.IsError() {
		return c.JSON(resp.StatusCode(), map[string]map[string]string{
			"data": {
				"message": "GitHub returned error: " + resp.String(),
			},
		})
	}
	if result.AccessToken == "" {
		return c.JSON(http.StatusInternalServerError, map[string]map[string]string{
			"data": {
				"message": "no access token received",
			},
		})
	}

	// Removed printing of access token to prevent sensitive credential exposure.

	return c.JSON(http.StatusOK, map[string]map[string]string{
		"data": {
			"message": "OAuth flow completed successfully. You are now being redirected.",
		},
	})
}
