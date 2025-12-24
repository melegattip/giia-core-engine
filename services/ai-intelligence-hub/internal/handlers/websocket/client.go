// Package websocket provides WebSocket functionality for real-time notifications.
package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/melegattip/giia-core-engine/pkg/logger"
)

// ErrUnauthorized is returned when authentication fails.
var ErrUnauthorized = fmt.Errorf("unauthorized")

// Client represents a WebSocket client connection.
type Client struct {
	hub *Hub

	// The WebSocket connection.
	conn *websocket.Conn

	// User ID associated with this client.
	userID uuid.UUID

	// Organization ID associated with this client.
	orgID uuid.UUID

	// Buffered channel of outbound messages.
	send chan []byte

	// Logger
	logger logger.Logger

	// Connection established time.
	connectedAt time.Time

	// Last activity time.
	lastActivity time.Time
}

// NewClient creates a new WebSocket client.
func NewClient(hub *Hub, conn *websocket.Conn, userID, orgID uuid.UUID, logger logger.Logger) *Client {
	now := time.Now()
	return &Client{
		hub:          hub,
		conn:         conn,
		userID:       userID,
		orgID:        orgID,
		send:         make(chan []byte, sendChannelSize),
		logger:       logger,
		connectedAt:  now,
		lastActivity: now,
	}
}

// ClientMessage represents a message received from a client.
type ClientMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// AckMessage represents an acknowledgment message.
type AckMessage struct {
	NotificationID uuid.UUID `json:"notification_id"`
}

// ReadPump pumps messages from the WebSocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.lastActivity = time.Now()
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error(nil, err, "WebSocket read error", logger.Tags{
					"user_id": c.userID.String(),
				})
			}
			break
		}

		c.lastActivity = time.Now()
		c.handleMessage(message)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming client messages.
func (c *Client) handleMessage(message []byte) {
	var msg ClientMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		c.sendError("invalid message format")
		return
	}

	switch msg.Type {
	case "ping":
		c.sendPong()

	case "ack":
		// Handle notification acknowledgment
		var ack AckMessage
		if err := json.Unmarshal(msg.Payload, &ack); err != nil {
			c.sendError("invalid ack payload")
			return
		}
		c.handleAck(ack)

	case "subscribe":
		// Client can subscribe to specific notification types
		c.handleSubscribe(msg.Payload)

	default:
		c.sendError("unknown message type")
	}
}

func (c *Client) handleAck(ack AckMessage) {
	c.logger.Debug(nil, "Notification acknowledged", logger.Tags{
		"notification_id": ack.NotificationID.String(),
		"user_id":         c.userID.String(),
	})
	// In production, update notification delivery status
}

func (c *Client) handleSubscribe(payload json.RawMessage) {
	c.logger.Debug(nil, "Subscription request received", logger.Tags{
		"user_id": c.userID.String(),
	})
	// In production, handle subscription filtering
}

func (c *Client) sendPong() {
	response := WebSocketMessage{
		Type:      "pong",
		Timestamp: time.Now(),
	}

	data, _ := json.Marshal(response)

	select {
	case c.send <- data:
	default:
		// Buffer full
	}
}

func (c *Client) sendError(message string) {
	response := WebSocketMessage{
		Type:      "error",
		Timestamp: time.Now(),
		Error:     message,
	}

	data, _ := json.Marshal(response)

	select {
	case c.send <- data:
	default:
		// Buffer full
	}
}

// GetUserID returns the client's user ID.
func (c *Client) GetUserID() uuid.UUID {
	return c.userID
}

// GetOrgID returns the client's organization ID.
func (c *Client) GetOrgID() uuid.UUID {
	return c.orgID
}

// GetConnectedAt returns when the client connected.
func (c *Client) GetConnectedAt() time.Time {
	return c.connectedAt
}

// GetLastActivity returns the client's last activity time.
func (c *Client) GetLastActivity() time.Time {
	return c.lastActivity
}
