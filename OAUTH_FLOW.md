# Axon CI Server - OAuth Authentication Flow

This document describes the OAuth authentication flow implemented in the Axon CI Server.

## Overview

The server now supports complete GitHub OAuth authentication with automatic user account creation and session management.

## API Endpoints

### Authentication Flow

1. **GET /v1/users/OAuthURL**
   - Returns GitHub OAuth authorization URL
   - Client should redirect user to this URL to start OAuth flow

2. **GET /v1/users/OAuthCallback**
   - Handles GitHub OAuth callback
   - Exchanges authorization code for access token
   - Creates new user account or logs in existing user
   - Sets secure session cookie
   - Returns user information

### User Management

3. **GET /v1/users/me**
   - Returns current authenticated user information
   - Requires valid session cookie

4. **POST /v1/users/logout**
   - Invalidates current session
   - Clears session cookie

## User Account Handling

### New User Creation
- When a user completes OAuth for the first time, a new account is created
- User information is fetched from GitHub API using the access token
- Account includes: GitHub ID, login, name, email, avatar URL, profile URL

### Existing User Login
- If user already exists (matched by GitHub ID), they are logged in
- User profile information is updated with latest data from GitHub
- New session is created for each login

### Session Management
- Sessions expire after 24 hours
- HTTP-only cookies for security
- Session invalidation on logout
- Automatic cleanup of expired sessions

## Error Handling

The system handles various edge cases:
- Missing or invalid OAuth codes
- GitHub API failures
- Incomplete user profile information
- Expired or invalid sessions
- Network errors during GitHub API calls

## Security Features

- Access tokens are not exposed in API responses
- HTTP-only session cookies prevent XSS attacks
- Session expiration limits exposure time
- Secure session ID generation using crypto/rand
- Input validation and error handling

## Configuration

Required configuration in `config.yaml`:
```yaml
github:
  client_id: "your_github_app_client_id"
  client_secret: "your_github_app_client_secret"
  redirect_uri: "http://localhost:8000/v1/users/OAuthCallback"
  scope: "user:email"
  state: "random_state_string"
```

## Testing

Run the comprehensive test suite:
```bash
go test ./controllers/users/... -v
```

The test suite includes:
- Storage layer tests (user and session storage)
- Service layer tests (user management logic)
- HTTP endpoint tests (authentication flow)
- Mock GitHub API for reliable testing