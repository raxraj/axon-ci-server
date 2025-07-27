package users

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/raxraj/axon-ci-server/controllers"
	"github.com/spf13/viper"
)

// TestOAuthFlowIntegration tests the complete OAuth flow end-to-end
func TestOAuthFlowIntegration(t *testing.T) {
	// Setup viper config for test
	viper.Set("github.client_id", "test-client-id")
	viper.Set("github.client_secret", "test-client-secret")
	viper.Set("github.redirect_uri", "http://localhost:8000/v1/users/OAuthCallback")
	viper.Set("github.scope", "user:email")
	viper.Set("github.state", "test-state")

	// Create a fresh user service for this test
	userStorage := NewInMemoryUserStorage()
	sessionStorage := NewInMemorySessionStorage()
	githubClient := NewMockGitHubClient()
	testUserService := NewUserServiceWithGitHubClient(userStorage, sessionStorage, githubClient)

	// Mock GitHub user data
	githubUser := &controllers.GitHubUser{
		ID:        54321,
		Login:     "integrationuser",
		Name:      "Integration Test User",
		Email:     "integration@example.com",
		AvatarURL: "https://github.com/integrationuser.png",
		HTMLURL:   "https://github.com/integrationuser",
	}

	// Setup mock GitHub responses
	githubClient.AddUser("valid_access_token", githubUser)

	// Override the global userService for this test
	originalUserService := userService
	userService = testUserService
	defer func() { userService = originalUserService }()

	// Test 1: Get OAuth URL
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/users/OAuthURL", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := OAuthInitiate(c)
	if err != nil {
		t.Fatalf("OAuthInitiate failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	var oauthResponse controllers.SuccessResponse
	err = json.Unmarshal(rec.Body.Bytes(), &oauthResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal OAuth response: %v", err)
	}

	// Test 2: Simulate OAuth callback (creating new user)
	// Note: We need to manually call the service since we can't mock the GitHub token exchange in this test
	user, session, err := testUserService.CreateOrLoginUser("valid_access_token")
	if err != nil {
		t.Fatalf("Failed to create user via service: %v", err)
	}

	if user.Login != "integrationuser" {
		t.Fatalf("Expected login 'integrationuser', got '%s'", user.Login)
	}

	if session.UserID != user.ID {
		t.Fatal("Session should belong to the created user")
	}

	// Test 3: Use session to get current user (create mock session cookie)
	sessionCookie := &http.Cookie{
		Name:  "session",
		Value: session.ID,
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	req.AddCookie(sessionCookie)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = GetCurrentUser(c)
	if err != nil {
		t.Fatalf("GetCurrentUser failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	var currentUserResponse controllers.SuccessResponse
	err = json.Unmarshal(rec.Body.Bytes(), &currentUserResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal current user response: %v", err)
	}

	currentUserData, ok := currentUserResponse.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Current user response data should be a map")
	}

	currentUser, ok := currentUserData["user"].(map[string]interface{})
	if !ok {
		t.Fatal("Current user should be a map")
	}

	if currentUser["login"] != "integrationuser" {
		t.Fatalf("Expected current user login 'integrationuser', got '%s'", currentUser["login"])
	}

	// Test 4: OAuth callback for existing user (login via service)
	user2, session2, err := testUserService.CreateOrLoginUser("valid_access_token")
	if err != nil {
		t.Fatalf("Second CreateOrLoginUser failed: %v", err)
	}

	if user2.Login != "integrationuser" {
		t.Fatalf("Expected login 'integrationuser' on second login, got '%s'", user2.Login)
	}

	// Should have same user ID (same user)
	if user.ID != user2.ID {
		t.Fatal("User ID should be the same for existing user login")
	}

	// Should have different session ID (new session)
	if session.ID == session2.ID {
		t.Fatal("Should create a new session for each login")
	}

	// Test 5: Logout
	req = httptest.NewRequest(http.MethodPost, "/v1/users/logout", nil)
	req.AddCookie(sessionCookie)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = Logout(c)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for logout, got %d", rec.Code)
	}

	// Test 6: Try to use invalidated session
	req = httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	req.AddCookie(sessionCookie)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = GetCurrentUser(c)
	if err != nil {
		t.Fatalf("GetCurrentUser after logout failed: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401 after logout, got %d", rec.Code)
	}

	fmt.Println("✅ OAuth flow integration test completed successfully!")
}