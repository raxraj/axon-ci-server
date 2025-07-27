package users

import (
	"errors"
	"sync"

	"github.com/raxraj/axon-ci-server/controllers"
)

// UserStorage defines the interface for user storage operations
type UserStorage interface {
	CreateUser(user *controllers.User) error
	GetUserByGitHubID(githubID int64) (*controllers.User, error)
	GetUserByID(id string) (*controllers.User, error)
	UpdateUser(user *controllers.User) error
}

// SessionStorage defines the interface for session storage operations
type SessionStorage interface {
	CreateSession(session *controllers.Session) error
	GetSession(id string) (*controllers.Session, error)
	DeleteSession(id string) error
}

// InMemoryUserStorage implements UserStorage using in-memory storage
type InMemoryUserStorage struct {
	users       map[string]*controllers.User // userID -> User
	usersByGHID map[int64]*controllers.User   // githubID -> User
	mu          sync.RWMutex
}

// NewInMemoryUserStorage creates a new in-memory user storage
func NewInMemoryUserStorage() *InMemoryUserStorage {
	return &InMemoryUserStorage{
		users:       make(map[string]*controllers.User),
		usersByGHID: make(map[int64]*controllers.User),
	}
}

func (s *InMemoryUserStorage) CreateUser(user *controllers.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user already exists
	if _, exists := s.usersByGHID[user.GitHubID]; exists {
		return errors.New("user with this GitHub ID already exists")
	}

	s.users[user.ID] = user
	s.usersByGHID[user.GitHubID] = user
	return nil
}

func (s *InMemoryUserStorage) GetUserByGitHubID(githubID int64) (*controllers.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.usersByGHID[githubID]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *InMemoryUserStorage) GetUserByID(id string) (*controllers.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *InMemoryUserStorage) UpdateUser(user *controllers.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user exists
	if _, exists := s.users[user.ID]; !exists {
		return errors.New("user not found")
	}

	s.users[user.ID] = user
	s.usersByGHID[user.GitHubID] = user
	return nil
}

// InMemorySessionStorage implements SessionStorage using in-memory storage
type InMemorySessionStorage struct {
	sessions map[string]*controllers.Session // sessionID -> Session
	mu       sync.RWMutex
}

// NewInMemorySessionStorage creates a new in-memory session storage
func NewInMemorySessionStorage() *InMemorySessionStorage {
	return &InMemorySessionStorage{
		sessions: make(map[string]*controllers.Session),
	}
}

func (s *InMemorySessionStorage) CreateSession(session *controllers.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.ID] = session
	return nil
}

func (s *InMemorySessionStorage) GetSession(id string) (*controllers.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[id]
	if !exists {
		return nil, errors.New("session not found")
	}
	return session, nil
}

func (s *InMemorySessionStorage) DeleteSession(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
	return nil
}