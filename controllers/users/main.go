package users

import (
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/raxraj/axon-ci-server/config"
	"github.com/raxraj/axon-ci-server/controllers"
	"github.com/spf13/viper"
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
	return c.JSON(http.StatusOK, controllers.SuccessResponse{
		Data: map[string]interface{}{
			"authorization_url": authorizationUri,
		},
		Message: "Please visit this URL to authorize the application with GitHub.",
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
		return c.JSON(http.StatusBadRequest, controllers.ErrorResponse{
			Message:   "Missing code",
			Data:      nil,
			ErrorCode: nil,
		})
	}

	result := &tokenResp{}
	resp, err := config.RestClient.R().
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
		return c.JSON(http.StatusInternalServerError, controllers.ErrorResponse{
			Message:   "failed to get access token",
			Data:      nil,
			ErrorCode: nil,
		})
	}
	if resp.IsError() {
		return c.JSON(resp.StatusCode(), controllers.ErrorResponse{
			Message:   "GitHub returned error: " + resp.String(),
			Data:      nil,
			ErrorCode: nil,
		})
	}
	if result.AccessToken == "" {
		return c.JSON(http.StatusInternalServerError, controllers.ErrorResponse{
			Message:   "no access token received",
			Data:      nil,
			ErrorCode: nil,
		})
	}

	// Removed printing of access token to prevent sensitive credential exposure.

	return c.JSON(http.StatusOK, controllers.SuccessResponse{
		Message: "OAuth flow completed successfully. You are now being redirected.",
		Data:    nil,
	})
}
