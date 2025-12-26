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

// AIHubClient provides methods to interact with the AI Intelligence Hub.
type AIHubClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAIHubClient creates a new AIHubClient.
func NewAIHubClient(baseURL string) *AIHubClient {
	return &AIHubClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Notification represents a notification from the AI Hub.
type Notification struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	OrganizationID string                 `json:"organization_id"`
	Type           string                 `json:"type"`
	Title          string                 `json:"title"`
	Message        string                 `json:"message"`
	Severity       string                 `json:"severity"`
	Category       string                 `json:"category"`
	Metadata       map[string]interface{} `json:"metadata"`
	IsRead         bool                   `json:"is_read"`
	ReadAt         *time.Time             `json:"read_at,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	ExpiresAt      *time.Time             `json:"expires_at,omitempty"`
}

// NotificationPreferences represents user notification preferences.
type NotificationPreferences struct {
	UserID             string   `json:"user_id"`
	OrganizationID     string   `json:"organization_id"`
	EmailEnabled       bool     `json:"email_enabled"`
	PushEnabled        bool     `json:"push_enabled"`
	WebSocketEnabled   bool     `json:"websocket_enabled"`
	EnabledCategories  []string `json:"enabled_categories"`
	DisabledCategories []string `json:"disabled_categories"`
	QuietHoursStart    string   `json:"quiet_hours_start"`
	QuietHoursEnd      string   `json:"quiet_hours_end"`
	MinimumSeverity    string   `json:"minimum_severity"`
}

// ListNotificationsResponse represents the response for listing notifications.
type ListNotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
	Total         int            `json:"total"`
	Page          int            `json:"page"`
	PageSize      int            `json:"page_size"`
	UnreadCount   int            `json:"unread_count"`
}

// UnreadCountResponse represents the unread notification count response.
type UnreadCountResponse struct {
	Count int `json:"count"`
}

// UpdateNotificationRequest represents a request to update a notification.
type UpdateNotificationRequest struct {
	IsRead bool `json:"is_read"`
}

// UpdatePreferencesRequest represents a request to update notification preferences.
type UpdatePreferencesRequest struct {
	EmailEnabled       *bool    `json:"email_enabled,omitempty"`
	PushEnabled        *bool    `json:"push_enabled,omitempty"`
	WebSocketEnabled   *bool    `json:"websocket_enabled,omitempty"`
	EnabledCategories  []string `json:"enabled_categories,omitempty"`
	DisabledCategories []string `json:"disabled_categories,omitempty"`
	QuietHoursStart    string   `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd      string   `json:"quiet_hours_end,omitempty"`
	MinimumSeverity    string   `json:"minimum_severity,omitempty"`
}

// ListNotifications lists notifications for a user.
func (c *AIHubClient) ListNotifications(ctx context.Context, userID, organizationID, accessToken string, page, pageSize int) (*ListNotificationsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/notifications?page=%d&page_size=%d", c.baseURL, page, pageSize)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("X-User-ID", userID)
	httpReq.Header.Set("X-Organization-ID", organizationID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("list notifications failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result ListNotificationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetNotification gets a notification by ID.
func (c *AIHubClient) GetNotification(ctx context.Context, notificationID, userID, organizationID, accessToken string) (*Notification, error) {
	url := fmt.Sprintf("%s/api/v1/notifications/%s", c.baseURL, notificationID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("X-User-ID", userID)
	httpReq.Header.Set("X-Organization-ID", organizationID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("get notification failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Notification Notification `json:"notification"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Notification, nil
}

// GetUnreadCount gets the unread notification count for a user.
func (c *AIHubClient) GetUnreadCount(ctx context.Context, userID, organizationID, accessToken string) (int, error) {
	url := c.baseURL + "/api/v1/notifications/unread-count"

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("X-User-ID", userID)
	httpReq.Header.Set("X-Organization-ID", organizationID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return 0, fmt.Errorf("get unread count failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result UnreadCountResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Count, nil
}

// MarkAsRead marks a notification as read.
func (c *AIHubClient) MarkAsRead(ctx context.Context, notificationID, userID, organizationID, accessToken string) error {
	body, err := json.Marshal(UpdateNotificationRequest{IsRead: true})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/notifications/%s", c.baseURL, notificationID)
	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("X-User-ID", userID)
	httpReq.Header.Set("X-Organization-ID", organizationID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("mark as read failed with status %d: %v", resp.StatusCode, errResp)
	}

	return nil
}

// DeleteNotification deletes a notification.
func (c *AIHubClient) DeleteNotification(ctx context.Context, notificationID, userID, organizationID, accessToken string) error {
	url := fmt.Sprintf("%s/api/v1/notifications/%s", c.baseURL, notificationID)
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("X-User-ID", userID)
	httpReq.Header.Set("X-Organization-ID", organizationID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("delete notification failed with status %d: %v", resp.StatusCode, errResp)
	}

	return nil
}

// GetPreferences gets notification preferences for a user.
func (c *AIHubClient) GetPreferences(ctx context.Context, userID, organizationID, accessToken string) (*NotificationPreferences, error) {
	url := c.baseURL + "/api/v1/notifications/preferences"

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("X-User-ID", userID)
	httpReq.Header.Set("X-Organization-ID", organizationID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("get preferences failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Preferences NotificationPreferences `json:"preferences"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Preferences, nil
}

// UpdatePreferences updates notification preferences for a user.
func (c *AIHubClient) UpdatePreferences(ctx context.Context, req UpdatePreferencesRequest, userID, organizationID, accessToken string) (*NotificationPreferences, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", c.baseURL+"/api/v1/notifications/preferences", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("X-User-ID", userID)
	httpReq.Header.Set("X-Organization-ID", organizationID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("update preferences failed with status %d: %v", resp.StatusCode, errResp)
	}

	var result struct {
		Preferences NotificationPreferences `json:"preferences"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Preferences, nil
}

// HealthCheck checks if the AI Hub service is healthy.
func (c *AIHubClient) HealthCheck(ctx context.Context) error {
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
