// Package websocket provides WebSocket functionality for real-time notifications.
package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 8192

	// Buffer size for client channels.
	sendChannelSize = 256

	// Maximum missed notifications to send on reconnect.
	maxMissedNotifications = 100
)

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	// Registered clients by user ID
	clients map[uuid.UUID]map[*Client]bool

	// Registered clients by organization ID
	orgClients map[uuid.UUID]map[*Client]bool

	// Register requests from clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Broadcast channel for notifications.
	broadcast chan *BroadcastMessage

	// Client last seen tracking for reconnection.
	lastSeen map[uuid.UUID]time.Time

	// Notification repository for fetching missed notifications.
	notifRepo providers.NotificationRepository

	// Logger
	logger logger.Logger

	// WebSocket upgrader.
	upgrader websocket.Upgrader

	// Mutex for thread-safe operations.
	mu sync.RWMutex

	// Metrics
	metrics *HubMetrics
}

// HubMetrics tracks WebSocket hub performance metrics.
type HubMetrics struct {
	TotalConnections  int64
	ActiveConnections int64
	MessagesSent      int64
	MessagesDropped   int64
	ReconnectionsSent int64
	AverageLatencyMs  float64
	mu                sync.Mutex
}

// BroadcastMessage represents a message to broadcast to clients.
type BroadcastMessage struct {
	// Target user ID (nil for org-wide broadcast)
	UserID *uuid.UUID

	// Target organization ID
	OrganizationID uuid.UUID

	// The notification to send
	Notification *domain.AINotification

	// Message type
	Type string

	// Timestamp
	SentAt time.Time
}

// NewHub creates a new WebSocket hub.
func NewHub(notifRepo providers.NotificationRepository, logger logger.Logger) *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		orgClients: make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan *BroadcastMessage, 1024),
		lastSeen:   make(map[uuid.UUID]time.Time),
		notifRepo:  notifRepo,
		logger:     logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
		},
		metrics: &HubMetrics{},
	}
}

// Run starts the hub's main event loop.
func (h *Hub) Run(ctx context.Context) {
	h.logger.Info(ctx, "WebSocket hub started", nil)

	for {
		select {
		case <-ctx.Done():
			h.logger.Info(ctx, "WebSocket hub shutting down", nil)
			h.closeAllClients()
			return

		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// HandleWebSocket handles WebSocket upgrade requests.
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract authentication info
	userID, err := h.extractUserID(r)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to extract user ID", nil)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orgID, err := h.extractOrgID(r)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to extract organization ID", nil)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to upgrade to WebSocket", nil)
		return
	}

	// Check for last seen time for reconnection handling
	h.mu.RLock()
	lastSeen, hasLastSeen := h.lastSeen[userID]
	h.mu.RUnlock()

	// Create new client
	client := NewClient(h, conn, userID, orgID, h.logger)

	// Register client
	h.register <- client

	h.logger.Info(ctx, "WebSocket client connected", logger.Tags{
		"user_id": userID.String(),
		"org_id":  orgID.String(),
	})

	// Send missed notifications if reconnecting
	if hasLastSeen {
		go h.sendMissedNotifications(ctx, client, lastSeen)
	}

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}

// BroadcastNotification sends a notification to appropriate clients.
func (h *Hub) BroadcastNotification(notification *domain.AINotification) {
	message := &BroadcastMessage{
		UserID:         &notification.UserID,
		OrganizationID: notification.OrganizationID,
		Notification:   notification,
		Type:           "notification",
		SentAt:         time.Now(),
	}

	select {
	case h.broadcast <- message:
		// Message queued successfully
	default:
		// Channel full, log dropped message
		h.metrics.mu.Lock()
		h.metrics.MessagesDropped++
		h.metrics.mu.Unlock()
	}
}

// BroadcastToOrg sends a notification to all users in an organization.
func (h *Hub) BroadcastToOrg(orgID uuid.UUID, notification *domain.AINotification) {
	message := &BroadcastMessage{
		UserID:         nil,
		OrganizationID: orgID,
		Notification:   notification,
		Type:           "notification",
		SentAt:         time.Now(),
	}

	select {
	case h.broadcast <- message:
		// Message queued successfully
	default:
		// Channel full
		h.metrics.mu.Lock()
		h.metrics.MessagesDropped++
		h.metrics.mu.Unlock()
	}
}

// GetStats returns hub statistics.
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.metrics.mu.Lock()
	defer h.metrics.mu.Unlock()

	totalClients := 0
	for _, clients := range h.clients {
		totalClients += len(clients)
	}

	return map[string]interface{}{
		"active_connections": totalClients,
		"total_connections":  h.metrics.TotalConnections,
		"messages_sent":      h.metrics.MessagesSent,
		"messages_dropped":   h.metrics.MessagesDropped,
		"reconnections_sent": h.metrics.ReconnectionsSent,
		"average_latency_ms": h.metrics.AverageLatencyMs,
		"registered_users":   len(h.clients),
		"registered_orgs":    len(h.orgClients),
	}
}

// Internal methods

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Register by user ID
	if h.clients[client.userID] == nil {
		h.clients[client.userID] = make(map[*Client]bool)
	}
	h.clients[client.userID][client] = true

	// Register by organization ID
	if h.orgClients[client.orgID] == nil {
		h.orgClients[client.orgID] = make(map[*Client]bool)
	}
	h.orgClients[client.orgID][client] = true

	h.metrics.mu.Lock()
	h.metrics.TotalConnections++
	h.metrics.ActiveConnections++
	h.metrics.mu.Unlock()
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Store last seen time for reconnection handling
	h.lastSeen[client.userID] = time.Now()

	// Remove from user clients
	if clients, ok := h.clients[client.userID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.clients, client.userID)
		}
	}

	// Remove from org clients
	if clients, ok := h.orgClients[client.orgID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.orgClients, client.orgID)
		}
	}

	// Close client
	close(client.send)

	h.metrics.mu.Lock()
	h.metrics.ActiveConnections--
	h.metrics.mu.Unlock()
}

func (h *Hub) broadcastMessage(message *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	start := time.Now()
	sentCount := 0

	// Convert notification to JSON
	wsMessage := &WebSocketMessage{
		Type:      message.Type,
		Timestamp: message.SentAt,
		Data:      toNotificationPayload(message.Notification),
	}

	data, err := json.Marshal(wsMessage)
	if err != nil {
		return
	}

	// If targeting a specific user
	if message.UserID != nil {
		if clients, ok := h.clients[*message.UserID]; ok {
			for client := range clients {
				// Verify organization match
				if client.orgID == message.OrganizationID {
					select {
					case client.send <- data:
						sentCount++
					default:
						// Client buffer full
					}
				}
			}
		}
	} else {
		// Broadcast to all clients in the organization
		if clients, ok := h.orgClients[message.OrganizationID]; ok {
			for client := range clients {
				select {
				case client.send <- data:
					sentCount++
				default:
					// Client buffer full
				}
			}
		}
	}

	// Update metrics
	h.metrics.mu.Lock()
	h.metrics.MessagesSent += int64(sentCount)
	latency := time.Since(start).Milliseconds()
	h.metrics.AverageLatencyMs = (h.metrics.AverageLatencyMs + float64(latency)) / 2
	h.metrics.mu.Unlock()
}

func (h *Hub) sendMissedNotifications(ctx context.Context, client *Client, since time.Time) {
	if h.notifRepo == nil {
		return
	}

	// Use a slight buffer to avoid missing notifications
	since = since.Add(-5 * time.Second)

	filters := &providers.NotificationFilters{
		Limit: maxMissedNotifications,
	}

	notifications, err := h.notifRepo.List(ctx, client.userID, client.orgID, filters)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to fetch missed notifications", logger.Tags{
			"user_id": client.userID.String(),
		})
		return
	}

	// Filter to only notifications after last seen
	var missed []*domain.AINotification
	for _, n := range notifications {
		if n.CreatedAt.After(since) {
			missed = append(missed, n)
		}
	}

	if len(missed) == 0 {
		return
	}

	h.logger.Info(ctx, "Sending missed notifications", logger.Tags{
		"user_id": client.userID.String(),
		"count":   len(missed),
	})

	// Send each missed notification
	for _, n := range missed {
		wsMessage := &WebSocketMessage{
			Type:      "missed_notification",
			Timestamp: n.CreatedAt,
			Data:      toNotificationPayload(n),
		}

		data, err := json.Marshal(wsMessage)
		if err != nil {
			continue
		}

		select {
		case client.send <- data:
			h.metrics.mu.Lock()
			h.metrics.ReconnectionsSent++
			h.metrics.mu.Unlock()
		default:
			// Client buffer full
			return
		}
	}
}

func (h *Hub) closeAllClients() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, clients := range h.clients {
		for client := range clients {
			close(client.send)
		}
	}
}

func (h *Hub) extractUserID(r *http.Request) (uuid.UUID, error) {
	// Try header first
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		// Try query parameter
		userIDStr = r.URL.Query().Get("user_id")
	}
	if userIDStr == "" {
		// Try to extract from JWT token
		// In production, implement proper JWT parsing
		return uuid.Nil, ErrUnauthorized
	}
	return uuid.Parse(userIDStr)
}

func (h *Hub) extractOrgID(r *http.Request) (uuid.UUID, error) {
	orgIDStr := r.Header.Get("X-Organization-ID")
	if orgIDStr == "" {
		orgIDStr = r.URL.Query().Get("organization_id")
	}
	if orgIDStr == "" {
		return uuid.Nil, ErrUnauthorized
	}
	return uuid.Parse(orgIDStr)
}

// WebSocketMessage represents a message sent over WebSocket.
type WebSocketMessage struct {
	Type      string               `json:"type"`
	Timestamp time.Time            `json:"timestamp"`
	Data      *NotificationPayload `json:"data,omitempty"`
	Error     string               `json:"error,omitempty"`
}

// NotificationPayload represents notification data in WebSocket messages.
type NotificationPayload struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	UserID         uuid.UUID `json:"user_id"`
	Type           string    `json:"type"`
	Priority       string    `json:"priority"`
	Title          string    `json:"title"`
	Summary        string    `json:"summary"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

func toNotificationPayload(n *domain.AINotification) *NotificationPayload {
	if n == nil {
		return nil
	}
	return &NotificationPayload{
		ID:             n.ID,
		OrganizationID: n.OrganizationID,
		UserID:         n.UserID,
		Type:           string(n.Type),
		Priority:       string(n.Priority),
		Title:          n.Title,
		Summary:        n.Summary,
		Status:         string(n.Status),
		CreatedAt:      n.CreatedAt,
	}
}
