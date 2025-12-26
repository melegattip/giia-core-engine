// Package clients provides HTTP and gRPC clients for integration testing.
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AuthClient provides methods to interact with the Auth Service.
type AuthClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAuthClient creates a new AuthClient.
func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Phone          string `json:"phone"`
	OrganizationID string `json:"organization_id"`
}

// RegisterResponse represents a user registration response.
type RegisterResponse struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents a login response.
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	User         struct {
		ID             string `json:"id"`
		Email          string `json:"email"`
		FirstName      string `json:"first_name"`
		LastName       string `json:"last_name"`
		OrganizationID string `json:"organization_id"`
	} `json:"user"`
}

// TokenResponse represents token data.
type TokenResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

// Register creates a new user.
func (c *AuthClient) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/auth/register", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("registration failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// Login authenticates a user and returns tokens.
func (c *AuthClient) Login(ctx context.Context, email, password string) (*TokenResponse, error) {
	body, err := json.Marshal(LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/auth/login", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("login failed with status %d: %v", resp.StatusCode, errResp)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &TokenResponse{
		AccessToken:  loginResp.AccessToken,
		RefreshToken: loginResp.RefreshToken,
		ExpiresIn:    loginResp.ExpiresIn,
	}, nil
}

// RefreshToken refreshes an access token.
func (c *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	body, err := json.Marshal(map[string]string{
		"refresh_token": refreshToken,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/auth/refresh", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("token refresh failed with status %d: %v", resp.StatusCode, errResp)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &TokenResponse{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}

// ValidateToken validates an access token.
func (c *AuthClient) ValidateToken(ctx context.Context, accessToken string) (bool, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/auth/me", nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return false, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// Logout invalidates the user's session.
func (c *AuthClient) Logout(ctx context.Context, accessToken string) error {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/auth/logout", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("logout failed with status %d", resp.StatusCode)
	}

	return nil
}

// GetCurrentUser gets the current user's information.
func (c *AuthClient) GetCurrentUser(ctx context.Context, accessToken string) (*LoginResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/auth/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("get current user failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// HealthCheck checks if the auth service is healthy.
func (c *AuthClient) HealthCheck(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}
