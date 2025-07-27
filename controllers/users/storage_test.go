package users

import (
	"testing"
	"time"

	"github.com/raxraj/axon-ci-server/controllers"
)

func TestInMemoryUserStorage(t *testing.T) {
	storage := NewInMemoryUserStorage()

	// Test user creation
	user := &controllers.User{
		ID:        "test-id",
		GitHubID:  12345,
		Login:     "testuser",
		Name:      "Test User",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := storage.CreateUser(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test duplicate user creation should fail
	err = storage.CreateUser(user)
	if err == nil {
		t.Fatal("Expected error when creating duplicate user")
	}

	// Test get user by GitHub ID
	retrieved, err := storage.GetUserByGitHubID(12345)
	if err != nil {
		t.Fatalf("Failed to get user by GitHub ID: %v", err)
	}
	if retrieved.Login != "testuser" {
		t.Fatalf("Expected login 'testuser', got '%s'", retrieved.Login)
	}

	// Test get user by ID
	retrieved, err = storage.GetUserByID("test-id")
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}
	if retrieved.Login != "testuser" {
		t.Fatalf("Expected login 'testuser', got '%s'", retrieved.Login)
	}

	// Test update user
	user.Name = "Updated Test User"
	err = storage.UpdateUser(user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	retrieved, err = storage.GetUserByID("test-id")
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}
	if retrieved.Name != "Updated Test User" {
		t.Fatalf("Expected name 'Updated Test User', got '%s'", retrieved.Name)
	}

	// Test get non-existent user
	_, err = storage.GetUserByGitHubID(99999)
	if err == nil {
		t.Fatal("Expected error when getting non-existent user")
	}

	_, err = storage.GetUserByID("non-existent")
	if err == nil {
		t.Fatal("Expected error when getting non-existent user")
	}
}

func TestInMemorySessionStorage(t *testing.T) {
	storage := NewInMemorySessionStorage()

	// Test session creation
	session := &controllers.Session{
		ID:        "session-id",
		UserID:    "user-id",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	err := storage.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test get session
	retrieved, err := storage.GetSession("session-id")
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}
	if retrieved.UserID != "user-id" {
		t.Fatalf("Expected user ID 'user-id', got '%s'", retrieved.UserID)
	}

	// Test delete session
	err = storage.DeleteSession("session-id")
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Test get deleted session should fail
	_, err = storage.GetSession("session-id")
	if err == nil {
		t.Fatal("Expected error when getting deleted session")
	}

	// Test get non-existent session
	_, err = storage.GetSession("non-existent")
	if err == nil {
		t.Fatal("Expected error when getting non-existent session")
	}
}