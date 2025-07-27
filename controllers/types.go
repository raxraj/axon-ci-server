package controllers

import "time"

type ErrorResponse struct {
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`       // Can be nil
	ErrorCode *int        `json:"error_code,omitempty"` // Optional error code for more specific error handling
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// User represents a user in the system
type User struct {
	ID          string    `json:"id"`
	GitHubID    int64     `json:"github_id"`
	Login       string    `json:"login"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	AvatarURL   string    `json:"avatar_url"`
	HTMLURL     string    `json:"html_url"`
	AccessToken string    `json:"-"` // Don't expose in JSON
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GitHubUser represents the user information from GitHub API
type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
