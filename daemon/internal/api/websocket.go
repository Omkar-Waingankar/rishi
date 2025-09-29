package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// WebSocket upgrader with CORS support
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Allow all origins for now (you may want to restrict this in production)
		return true
	},
}

// WebSocketConnection represents a WebSocket connection to an R client
type WebSocketConnection struct {
	conn   *websocket.Conn
	send   chan []byte
	client string
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	ID      string      `json:"id,omitempty"`
	Type    string      `json:"type"`
	Tool    string      `json:"tool,omitempty"`
	Command string      `json:"command,omitempty"`
	Input   interface{} `json:"input,omitempty"`
	Result  interface{} `json:"result,omitempty"`
}

// WebSocketManager manages WebSocket connections
type WebSocketManager struct {
	connections     map[string]*WebSocketConnection
	mu              sync.RWMutex
	pendingRequests map[string]interface{}
	requestMu       sync.RWMutex
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		connections:     make(map[string]*WebSocketConnection),
		pendingRequests: make(map[string]interface{}),
	}
}

// HandleWebSocket handles WebSocket connections
func (s *ServerClient) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}

	// Create WebSocket connection
	wsConn := &WebSocketConnection{
		conn: conn,
		send: make(chan []byte, 256),
	}

	// Handle the connection
	go wsConn.writePump()
	go wsConn.readPump(s)
}

// readPump pumps messages from the WebSocket connection
func (c *WebSocketConnection) readPump(s *ServerClient) {
	defer func() {
		c.conn.Close()
		if c.client != "" {
			s.wsManager.removeConnection(c.client)
		}
	}()

	c.conn.SetReadLimit(10 * 1024 * 1024) // 10MB limit
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Msg("WebSocket error")
			}
			break
		}

		// Parse message
		var wsMsg WebSocketMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Error().Err(err).Msg("Failed to parse WebSocket message")
			continue
		}

		// Handle different message types
		switch wsMsg.Type {
		case "handshake":
			c.client = "r_tool_rpc"
			s.wsManager.addConnection(c.client, c)
			log.Info().Msg("R Tool RPC client connected via WebSocket")

		case "tool_response":
			// Handle response from R client
			if wsMsg.ID != "" && wsMsg.Result != nil {
				s.wsManager.handleToolResponse(wsMsg.ID, wsMsg.Tool, wsMsg.Command, wsMsg.Result)
			}
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *WebSocketConnection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Error().Err(err).Msg("Failed to write WebSocket message")
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// sendMessage sends a message to the WebSocket connection
func (c *WebSocketConnection) sendMessage(data []byte) {
	select {
	case c.send <- data:
	default:
		close(c.send)
	}
}

// addConnection adds a WebSocket connection to the manager
func (m *WebSocketManager) addConnection(clientID string, conn *WebSocketConnection) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[clientID] = conn
}

// removeConnection removes a WebSocket connection from the manager
func (m *WebSocketManager) removeConnection(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, clientID)
}

// sendToClient sends a message to a specific client
func (m *WebSocketManager) sendToClient(clientID string, message []byte) error {
	m.mu.RLock()
	conn, exists := m.connections[clientID]
	m.mu.RUnlock()

	if !exists {
		return ErrClientNotConnected
	}

	conn.sendMessage(message)
	return nil
}
