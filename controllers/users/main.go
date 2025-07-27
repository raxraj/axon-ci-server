package users

import (
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/raxraj/axon-ci-server/config"
	"github.com/raxraj/axon-ci-server/controllers"
	"github.com/spf13/viper"
)

var (
	userService *UserService
)

func init() {
	userStorage := NewInMemoryUserStorage()
	sessionStorage := NewInMemorySessionStorage()
	userService = NewUserService(userStorage, sessionStorage)
}

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

	// Exchange code for access token
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

	// Create or login user using the access token
	user, session, err := userService.CreateOrLoginUser(result.AccessToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, controllers.ErrorResponse{
			Message:   "failed to create or login user: " + err.Error(),
			Data:      nil,
			ErrorCode: nil,
		})
	}

	// Set session cookie for secure session management
	cookie := &http.Cookie{
		Name:     "session",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	}
	c.SetCookie(cookie)

	// Return success response with user info (without sensitive data)
	return c.JSON(http.StatusOK, controllers.SuccessResponse{
		Message: "OAuth flow completed successfully. User authenticated.",
		Data: map[string]interface{}{
			"user": map[string]interface{}{
				"id":         user.ID,
				"login":      user.Login,
				"name":       user.Name,
				"email":      user.Email,
				"avatar_url": user.AvatarURL,
				"html_url":   user.HTMLURL,
				"created_at": user.CreatedAt,
			},
			"session_expires_at": session.ExpiresAt,
		},
	})
}

// GetCurrentUser returns the current authenticated user
func GetCurrentUser(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, controllers.ErrorResponse{
			Message:   "no session found",
			Data:      nil,
			ErrorCode: nil,
		})
	}

	user, err := userService.GetUserBySession(cookie.Value)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, controllers.ErrorResponse{
			Message:   "invalid or expired session",
			Data:      nil,
			ErrorCode: nil,
		})
	}

	return c.JSON(http.StatusOK, controllers.SuccessResponse{
		Message: "User retrieved successfully",
		Data: map[string]interface{}{
			"user": map[string]interface{}{
				"id":         user.ID,
				"login":      user.Login,
				"name":       user.Name,
				"email":      user.Email,
				"avatar_url": user.AvatarURL,
				"html_url":   user.HTMLURL,
				"created_at": user.CreatedAt,
			},
		},
	})
}

// Logout invalidates the current session
func Logout(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil {
		return c.JSON(http.StatusOK, controllers.SuccessResponse{
			Message: "logged out successfully",
			Data:    nil,
		})
	}

	// Invalidate session
	userService.InvalidateSession(cookie.Value)

	// Clear the session cookie
	clearCookie := &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}
	c.SetCookie(clearCookie)

	return c.JSON(http.StatusOK, controllers.SuccessResponse{
		Message: "logged out successfully",
		Data:    nil,
	})
}
