package users

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/raxraj/axon-ci-server/controllers"
	"github.com/spf13/viper"
)

func TestOAuthInitiate(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/OAuthURL", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set up config
	viper.Set("github.client_id", "test-client-id")
	viper.Set("github.redirect_uri", "http://localhost:8000/auth/github/callback")
	viper.Set("github.scope", "user:email")
	viper.Set("github.state", "random-state")

	// Execute
	err := OAuthInitiate(c)
	if err != nil {
		t.Fatalf("OAuthInitiate failed: %v", err)
	}

	// Verify
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	var response controllers.SuccessResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Response data should be a map")
	}

	authURL, ok := data["authorization_url"].(string)
	if !ok {
		t.Fatal("authorization_url should be a string")
	}

	if !strings.Contains(authURL, "github.com/login/oauth/authorize") {
		t.Fatalf("Expected GitHub OAuth URL, got %s", authURL)
	}

	if !strings.Contains(authURL, "client_id=test-client-id") {
		t.Fatal("URL should contain client_id")
	}
}

func TestOAuthCallback_MissingCode(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/OAuthCallback", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := OAuthCallback(c)
	if err != nil {
		t.Fatalf("OAuthCallback failed: %v", err)
	}

	// Verify
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", rec.Code)
	}

	var response controllers.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Message != "Missing code" {
		t.Fatalf("Expected 'Missing code', got '%s'", response.Message)
	}
}

func TestGetCurrentUser_NoSession(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := GetCurrentUser(c)
	if err != nil {
		t.Fatalf("GetCurrentUser failed: %v", err)
	}

	// Verify
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401, got %d", rec.Code)
	}

	var response controllers.ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Message != "no session found" {
		t.Fatalf("Expected 'no session found', got '%s'", response.Message)
	}
}

func TestGetCurrentUser_ValidSession(t *testing.T) {
	// Create test user and session
	userStorage := NewInMemoryUserStorage()
	sessionStorage := NewInMemorySessionStorage()
	testUserService := NewUserService(userStorage, sessionStorage)

	// Create a test user
	user := &controllers.User{
		ID:       "test-user-id",
		GitHubID: 12345,
		Login:    "testuser",
		Name:     "Test User",
		Email:    "test@example.com",
	}
	userStorage.CreateUser(user)

	// Create a test session
	session, err := testUserService.CreateSession(user.ID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Override the global userService for this test
	originalUserService := userService
	userService = testUserService
	defer func() { userService = originalUserService }()

	// Setup request with session cookie
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: session.ID,
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err = GetCurrentUser(c)
	if err != nil {
		t.Fatalf("GetCurrentUser failed: %v", err)
	}

	// Verify
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	var response controllers.SuccessResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Response data should be a map")
	}

	userData, ok := data["user"].(map[string]interface{})
	if !ok {
		t.Fatal("User data should be a map")
	}

	if userData["login"] != "testuser" {
		t.Fatalf("Expected login 'testuser', got '%s'", userData["login"])
	}
}

func TestLogout(t *testing.T) {
	// Create test user and session
	userStorage := NewInMemoryUserStorage()
	sessionStorage := NewInMemorySessionStorage()
	testUserService := NewUserService(userStorage, sessionStorage)

	// Create a test session
	session, err := testUserService.CreateSession("test-user-id")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Override the global userService for this test
	originalUserService := userService
	userService = testUserService
	defer func() { userService = originalUserService }()

	// Setup request with session cookie
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/users/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: session.ID,
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err = Logout(c)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Verify
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	var response controllers.SuccessResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Message != "logged out successfully" {
		t.Fatalf("Expected 'logged out successfully', got '%s'", response.Message)
	}

	// Verify session was invalidated
	_, err = sessionStorage.GetSession(session.ID)
	if err == nil {
		t.Fatal("Session should be invalidated after logout")
	}

	// Check that the cookie was cleared
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			sessionCookie = cookie
			break
		}
	}

	if sessionCookie == nil {
		t.Fatal("Session cookie should be present in response")
	}

	if sessionCookie.Value != "" {
		t.Fatal("Session cookie should be cleared")
	}

	if sessionCookie.MaxAge != -1 {
		t.Fatal("Session cookie should have MaxAge=-1")
	}
}