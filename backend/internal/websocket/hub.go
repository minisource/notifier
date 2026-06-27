package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/minisource/go-common/logging"
	"github.com/minisource/notifier/internal/models"
)

// Client represents a WebSocket client
type Client struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	Conn     *websocket.Conn
	Hub      *Hub
	Send     chan []byte
	mu       sync.Mutex
	lastPing time.Time
}

// Hub maintains active WebSocket connections and broadcasts notifications
type Hub struct {
	clients    map[uuid.UUID]*Client
	userIndex  map[uuid.UUID]map[uuid.UUID]*Client // userID -> map of client IDs
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	mu         sync.RWMutex
	logger     logging.Logger
	ctx        context.Context
	cancel     context.CancelFunc
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	UserID       *uuid.UUID // nil for broadcast to all
	Notification *models.Notification
}

// NotificationMessage represents the WebSocket message format
type NotificationMessage struct {
	Type string               `json:"type"`
	Data *models.Notification `json:"data"`
	Time time.Time            `json:"time"`
}

// NewHub creates a new WebSocket hub
func NewHub(logger logging.Logger) *Hub {
	ctx, cancel := context.WithCancel(context.Background())

	return &Hub{
		clients:    make(map[uuid.UUID]*Client),
		userIndex:  make(map[uuid.UUID]map[uuid.UUID]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage, 256),
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start starts the hub
func (h *Hub) Start() {
	h.logger.Info(logging.General, logging.Startup, "Starting WebSocket hub", nil)

	go h.run()
	go h.pingClients()

	h.logger.Info(logging.General, logging.Startup, "WebSocket hub started successfully", nil)
}

// Stop stops the hub
func (h *Hub) Stop() {
	h.logger.Info(logging.General, logging.Startup, "Stopping WebSocket hub", nil)

	h.cancel()
	close(h.register)
	close(h.unregister)
	close(h.broadcast)

	// Close all client connections
	h.mu.Lock()
	for _, client := range h.clients {
		close(client.Send)
		client.Conn.Close()
	}
	h.mu.Unlock()

	h.logger.Info(logging.General, logging.Startup, "WebSocket hub stopped successfully", nil)
}

// run handles hub operations
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		case <-h.ctx.Done():
			h.logger.Debug(logging.General, logging.Startup, "Hub run loop stopped", nil)
			return
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client

	// Index by user ID
	if _, ok := h.userIndex[client.UserID]; !ok {
		h.userIndex[client.UserID] = make(map[uuid.UUID]*Client)
	}
	h.userIndex[client.UserID][client.ID] = client

	h.logger.Debug(logging.Internal, logging.Api, "Client registered", map[logging.ExtraKey]interface{}{
		"clientId":     client.ID,
		"userId":       client.UserID,
		"totalClients": len(h.clients),
	})
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)

		// Remove from user index
		if userClients, ok := h.userIndex[client.UserID]; ok {
			delete(userClients, client.ID)
			if len(userClients) == 0 {
				delete(h.userIndex, client.UserID)
			}
		}

		close(client.Send)

		h.logger.Debug(logging.Internal, logging.Api, "Client unregistered", map[logging.ExtraKey]interface{}{
			"clientId":     client.ID,
			"userId":       client.UserID,
			"totalClients": len(h.clients),
		})
	}
}

// broadcastMessage broadcasts a message to relevant clients
func (h *Hub) broadcastMessage(message *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Serialize notification
	notifMsg := &NotificationMessage{
		Type: "notification",
		Data: message.Notification,
		Time: time.Now(),
	}

	data, err := json.Marshal(notifMsg)
	if err != nil {
		h.logger.Error(logging.Internal, logging.Api, "Failed to marshal notification", map[logging.ExtraKey]interface{}{
			"error": err.Error(),
		})
		return
	}

	if message.UserID != nil {
		// Send to specific user's clients
		if userClients, ok := h.userIndex[*message.UserID]; ok {
			h.logger.Debug(logging.Internal, logging.Api, "Broadcasting to user", map[logging.ExtraKey]interface{}{
				"userId":      *message.UserID,
				"clientCount": len(userClients),
			})

			for _, client := range userClients {
				select {
				case client.Send <- data:
				default:
					h.logger.Warn(logging.Internal, logging.Api, "Client send buffer full", map[logging.ExtraKey]interface{}{
						"clientId": client.ID,
					})
				}
			}
		}
	} else {
		// Broadcast to all clients
		h.logger.Debug(logging.Internal, logging.Api, "Broadcasting to all clients", map[logging.ExtraKey]interface{}{
			"clientCount": len(h.clients),
		})

		for _, client := range h.clients {
			select {
			case client.Send <- data:
			default:
				h.logger.Warn(logging.Internal, logging.Api, "Client send buffer full", map[logging.ExtraKey]interface{}{
					"clientId": client.ID,
				})
			}
		}
	}
}

// BroadcastToUser sends a notification to a specific user
func (h *Hub) BroadcastToUser(userID uuid.UUID, notification *models.Notification) {
	h.logger.Debug(logging.Internal, logging.Api, "Queueing notification for user", map[logging.ExtraKey]interface{}{
		"userId":         userID,
		"notificationId": notification.ID,
	})

	select {
	case h.broadcast <- &BroadcastMessage{
		UserID:       &userID,
		Notification: notification,
	}:
	default:
		h.logger.Warn(logging.Internal, logging.Api, "Broadcast channel full", map[logging.ExtraKey]interface{}{
			"userId": userID,
		})
	}
}

// BroadcastToAll sends a notification to all connected clients
func (h *Hub) BroadcastToAll(notification *models.Notification) {
	h.logger.Debug(logging.Internal, logging.Api, "Queueing notification for all", map[logging.ExtraKey]interface{}{
		"notificationId": notification.ID,
	})

	select {
	case h.broadcast <- &BroadcastMessage{
		Notification: notification,
	}:
	default:
		h.logger.Warn(logging.Internal, logging.Api, "Broadcast channel full", nil)
	}
}

// GetConnectedUsers returns the count of connected users
func (h *Hub) GetConnectedUsers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.userIndex)
}

// GetTotalConnections returns the total number of connections
func (h *Hub) GetTotalConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// pingClients periodically pings clients to keep connections alive
func (h *Hub) pingClients() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.mu.RLock()
			clients := make([]*Client, 0, len(h.clients))
			for _, client := range h.clients {
				clients = append(clients, client)
			}
			h.mu.RUnlock()

			for _, client := range clients {
				client.mu.Lock()
				if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					h.logger.Debug(logging.Internal, logging.Api, "Failed to ping client", map[logging.ExtraKey]interface{}{
						"clientId": client.ID,
						"error":    err.Error(),
					})
					client.mu.Unlock()
					h.unregister <- client
					continue
				}
				client.lastPing = time.Now()
				client.mu.Unlock()
			}
		case <-h.ctx.Done():
			return
		}
	}
}

// HandleWebSocket handles WebSocket upgrade and connection
func (h *Hub) HandleWebSocket() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		// Get user ID from query or header
		userIDStr := c.Query("userId")
		if userIDStr == "" {
			h.logger.Warn(logging.Internal, logging.Api, "WebSocket connection without userId", nil)
			c.Close()
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			h.logger.Warn(logging.Internal, logging.Api, "Invalid userId in WebSocket connection", map[logging.ExtraKey]interface{}{})
			c.Close()
			return
		}

		// Create new client
		client := &Client{
			ID:       uuid.New(),
			UserID:   userID,
			Conn:     c,
			Hub:      h,
			Send:     make(chan []byte, 256),
			lastPing: time.Now(),
		}

		// Register client
		h.register <- client

		h.logger.Info(logging.Internal, logging.Api, "WebSocket client connected", map[logging.ExtraKey]interface{}{
			"clientId": client.ID,
			"userId":   userID,
		})

		// Start read and write pumps
		go client.writePump()
		client.readPump()
	})
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		messageType, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.logger.Error(logging.Internal, logging.Api, "WebSocket read error", map[logging.ExtraKey]interface{}{
					"error": err.Error(),
				})
			}
			break
		}

		c.Hub.logger.Debug(logging.Internal, logging.Api, "Received WebSocket message", map[logging.ExtraKey]interface{}{
			"clientId":    c.ID,
			"messageType": messageType,
			"message":     string(message),
		})

		// Handle client messages if needed (e.g., acknowledgments)
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.mu.Lock()
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			c.mu.Unlock()

			if err != nil {
				c.Hub.logger.Error(logging.Internal, logging.Api, "Failed to write message", map[logging.ExtraKey]interface{}{
					"clientId": c.ID,
					"error":    err.Error(),
				})
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			c.mu.Lock()
			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			c.mu.Unlock()

			if err != nil {
				return
			}
		}
	}
}
