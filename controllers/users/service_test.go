package users

import (
	"fmt"
	"testing"
	"time"

	"github.com/raxraj/axon-ci-server/controllers"
)

// MockGitHubClient for testing
type MockGitHubClient struct {
	users map[string]*controllers.GitHubUser
}

func NewMockGitHubClient() *MockGitHubClient {
	return &MockGitHubClient{
		users: make(map[string]*controllers.GitHubUser),
	}
}

func (m *MockGitHubClient) AddUser(token string, user *controllers.GitHubUser) {
	m.users[token] = user
}

func (m *MockGitHubClient) GetUser(accessToken string) (*controllers.GitHubUser, error) {
	user, exists := m.users[accessToken]
	if !exists {
		return nil, fmt.Errorf("invalid token")
	}
	return user, nil
}

func TestUserService_CreateOrLoginUser(t *testing.T) {
	// Create test instances
	userStorage := NewInMemoryUserStorage()
	sessionStorage := NewInMemorySessionStorage()
	githubClient := NewMockGitHubClient()
	service := NewUserServiceWithGitHubClient(userStorage, sessionStorage, githubClient)

	// Setup mock GitHub user
	githubUser := &controllers.GitHubUser{
		ID:        12345,
		Login:     "testuser",
		Name:      "Test User",
		Email:     "test@example.com",
		AvatarURL: "https://github.com/testuser.png",
		HTMLURL:   "https://github.com/testuser",
	}
	githubClient.AddUser("valid-token", githubUser)

	// Test creating a new user
	user, session, err := service.CreateOrLoginUser("valid-token")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.GitHubID != 12345 {
		t.Fatalf("Expected GitHub ID 12345, got %d", user.GitHubID)
	}
	if user.Login != "testuser" {
		t.Fatalf("Expected login 'testuser', got '%s'", user.Login)
	}
	if session.UserID != user.ID {
		t.Fatalf("Session user ID should match user ID")
	}

	// Test logging in existing user
	user2, session2, err := service.CreateOrLoginUser("valid-token")
	if err != nil {
		t.Fatalf("Failed to login existing user: %v", err)
	}

	if user2.ID != user.ID {
		t.Fatalf("Should return the same user for existing user")
	}
	if session2.ID == session.ID {
		t.Fatalf("Should create a new session for each login")
	}

	// Test with invalid token
	_, _, err = service.CreateOrLoginUser("invalid-token")
	if err == nil {
		t.Fatal("Expected error with invalid token")
	}

	// Test with incomplete GitHub user info
	incompleteUser := &controllers.GitHubUser{
		ID:    0, // Missing ID
		Login: "",
	}
	githubClient.AddUser("incomplete-token", incompleteUser)

	_, _, err = service.CreateOrLoginUser("incomplete-token")
	if err == nil {
		t.Fatal("Expected error with incomplete user info")
	}
}

func TestUserService_GetUserBySession(t *testing.T) {
	userStorage := NewInMemoryUserStorage()
	sessionStorage := NewInMemorySessionStorage()
	githubClient := NewMockGitHubClient()
	service := NewUserServiceWithGitHubClient(userStorage, sessionStorage, githubClient)

	// Create test user
	user := &controllers.User{
		ID:        "test-user-id",
		GitHubID:  12345,
		Login:     "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	userStorage.CreateUser(user)

	// Create test session
	session := &controllers.Session{
		ID:        "test-session-id",
		UserID:    user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	sessionStorage.CreateSession(session)

	// Test valid session
	retrievedUser, err := service.GetUserBySession("test-session-id")
	if err != nil {
		t.Fatalf("Failed to get user by session: %v", err)
	}
	if retrievedUser.ID != user.ID {
		t.Fatalf("Expected user ID '%s', got '%s'", user.ID, retrievedUser.ID)
	}

	// Test expired session
	expiredSession := &controllers.Session{
		ID:        "expired-session-id",
		UserID:    user.ID,
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	sessionStorage.CreateSession(expiredSession)

	_, err = service.GetUserBySession("expired-session-id")
	if err == nil {
		t.Fatal("Expected error for expired session")
	}

	// Test non-existent session
	_, err = service.GetUserBySession("non-existent-session")
	if err == nil {
		t.Fatal("Expected error for non-existent session")
	}
}

func TestUserService_InvalidateSession(t *testing.T) {
	userStorage := NewInMemoryUserStorage()
	sessionStorage := NewInMemorySessionStorage()
	githubClient := NewMockGitHubClient()
	service := NewUserServiceWithGitHubClient(userStorage, sessionStorage, githubClient)

	// Create test session
	session := &controllers.Session{
		ID:        "test-session-id",
		UserID:    "test-user-id",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}
	sessionStorage.CreateSession(session)

	// Verify session exists
	_, err := sessionStorage.GetSession("test-session-id")
	if err != nil {
		t.Fatalf("Session should exist before invalidation")
	}

	// Invalidate session
	err = service.InvalidateSession("test-session-id")
	if err != nil {
		t.Fatalf("Failed to invalidate session: %v", err)
	}

	// Verify session is gone
	_, err = sessionStorage.GetSession("test-session-id")
	if err == nil {
		t.Fatal("Session should not exist after invalidation")
	}
}

func TestUserService_CreateSession(t *testing.T) {
	userStorage := NewInMemoryUserStorage()
	sessionStorage := NewInMemorySessionStorage()
	githubClient := NewMockGitHubClient()
	service := NewUserServiceWithGitHubClient(userStorage, sessionStorage, githubClient)

	session, err := service.CreateSession("test-user-id")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.UserID != "test-user-id" {
		t.Fatalf("Expected user ID 'test-user-id', got '%s'", session.UserID)
	}

	if session.ID == "" {
		t.Fatal("Session ID should not be empty")
	}

	if session.ExpiresAt.Before(time.Now()) {
		t.Fatal("Session should not be expired upon creation")
	}

	// Verify session was stored
	_, err = sessionStorage.GetSession(session.ID)
	if err != nil {
		t.Fatal("Session should be stored in storage")
	}
}