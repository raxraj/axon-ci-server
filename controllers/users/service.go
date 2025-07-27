package users

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/raxraj/axon-ci-server/config"
	"github.com/raxraj/axon-ci-server/controllers"
)

// GitHubClient defines the interface for GitHub API interactions
type GitHubClient interface {
	GetUser(accessToken string) (*controllers.GitHubUser, error)
}

// DefaultGitHubClient implements GitHubClient using the real GitHub API
type DefaultGitHubClient struct{}

func (c *DefaultGitHubClient) GetUser(accessToken string) (*controllers.GitHubUser, error) {
	result := &controllers.GitHubUser{}
	resp, err := config.RestClient.R().
		SetHeader("Authorization", "Bearer "+accessToken).
		SetHeader("Accept", "application/vnd.github.v3+json").
		SetResult(result).
		Get("https://api.github.com/user")

	if err != nil {
		return nil, fmt.Errorf("failed to fetch user from GitHub: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("GitHub API error: %s", resp.String())
	}

	return result, nil
}

// UserService handles user-related business logic
type UserService struct {
	userStorage    UserStorage
	sessionStorage SessionStorage
	githubClient   GitHubClient
}

// NewUserService creates a new user service
func NewUserService(userStorage UserStorage, sessionStorage SessionStorage) *UserService {
	return &UserService{
		userStorage:    userStorage,
		sessionStorage: sessionStorage,
		githubClient:   &DefaultGitHubClient{},
	}
}

// NewUserServiceWithGitHubClient creates a new user service with a custom GitHub client
func NewUserServiceWithGitHubClient(userStorage UserStorage, sessionStorage SessionStorage, githubClient GitHubClient) *UserService {
	return &UserService{
		userStorage:    userStorage,
		sessionStorage: sessionStorage,
		githubClient:   githubClient,
	}
}

// CreateOrLoginUser creates a new user or logs in an existing user
func (s *UserService) CreateOrLoginUser(accessToken string) (*controllers.User, *controllers.Session, error) {
	// Fetch user info from GitHub
	githubUser, err := s.githubClient.GetUser(accessToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get GitHub user: %w", err)
	}

	// Validate required fields
	if githubUser.ID == 0 || githubUser.Login == "" {
		return nil, nil, fmt.Errorf("incomplete GitHub user information: missing ID or login")
	}

	// Check if user already exists
	existingUser, err := s.userStorage.GetUserByGitHubID(githubUser.ID)
	if err == nil {
		// User exists, update their information and create session
		existingUser.Name = githubUser.Name
		existingUser.Email = githubUser.Email
		existingUser.AvatarURL = githubUser.AvatarURL
		existingUser.HTMLURL = githubUser.HTMLURL
		existingUser.AccessToken = accessToken
		existingUser.UpdatedAt = time.Now()

		err = s.userStorage.UpdateUser(existingUser)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to update user: %w", err)
		}

		// Create session for existing user
		session, err := s.CreateSession(existingUser.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create session: %w", err)
		}

		return existingUser, session, nil
	}

	// User doesn't exist, create new user
	userID, err := generateID()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate user ID: %w", err)
	}

	newUser := &controllers.User{
		ID:          userID,
		GitHubID:    githubUser.ID,
		Login:       githubUser.Login,
		Name:        githubUser.Name,
		Email:       githubUser.Email,
		AvatarURL:   githubUser.AvatarURL,
		HTMLURL:     githubUser.HTMLURL,
		AccessToken: accessToken,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.userStorage.CreateUser(newUser)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create session for new user
	session, err := s.CreateSession(newUser.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	return newUser, session, nil
}

// CreateSession creates a new session for a user
func (s *UserService) CreateSession(userID string) (*controllers.Session, error) {
	sessionID, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	session := &controllers.Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours session
	}

	err = s.sessionStorage.CreateSession(session)
	if err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return session, nil
}

// GetUserBySession retrieves user by session ID
func (s *UserService) GetUserBySession(sessionID string) (*controllers.User, error) {
	session, err := s.sessionStorage.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		s.sessionStorage.DeleteSession(sessionID)
		return nil, fmt.Errorf("session expired")
	}

	user, err := s.userStorage.GetUserByID(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

// InvalidateSession removes a session
func (s *UserService) InvalidateSession(sessionID string) error {
	return s.sessionStorage.DeleteSession(sessionID)
}

// generateID generates a random hex string for IDs
func generateID() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}