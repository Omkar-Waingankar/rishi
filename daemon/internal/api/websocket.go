package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
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
	mu     sync.Mutex
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
	// Check authorization
	auth := r.Header.Get("Authorization")
	if auth != "Bearer "+s.toolRPCToken {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

// sendToolCommand sends a tool command to the R client and waits for response
func (m *WebSocketManager) sendToolCommand(tool, command string, input interface{}, responseChan interface{}) (string, error) {
	// Generate unique ID for this request
	requestID := generateRequestID()

	// Store response channel
	m.requestMu.Lock()
	m.pendingRequests[requestID] = responseChan
	m.requestMu.Unlock()

	// Create message with general structure
	msg := WebSocketMessage{
		ID:      requestID,
		Type:    "tool_request",
		Tool:    tool,
		Command: command,
		Input:   input,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		// Clean up on error
		m.requestMu.Lock()
		delete(m.pendingRequests, requestID)
		m.requestMu.Unlock()
		return "", err
	}

	// Send to R client
	if err := m.sendToClient("r_tool_rpc", data); err != nil {
		// Clean up on error
		m.requestMu.Lock()
		delete(m.pendingRequests, requestID)
		m.requestMu.Unlock()
		return "", err
	}

	return requestID, nil
}

// sendTextEditorCommand sends a text editor command to the R client and waits for response
func (m *WebSocketManager) sendTextEditorCommand(input textEditorViewInput) (*textEditorViewOutput, error) {
	// Create response channel
	responseChan := make(chan *textEditorViewOutput, 1)

	// Send the tool command
	requestID, err := m.sendToolCommand("text_editor", "view", input, responseChan)
	if err != nil {
		close(responseChan)
		return nil, err
	}

	// Clean up function
	cleanup := func() {
		m.requestMu.Lock()
		delete(m.pendingRequests, requestID)
		m.requestMu.Unlock()
		close(responseChan)
	}

	// Wait for response with timeout
	select {
	case response := <-responseChan:
		cleanup()
		return response, nil
	case <-time.After(30 * time.Second):
		cleanup()
		return nil, ErrTimeout
	}
}

// handleToolResponse handles a response from the R client for any tool
func (m *WebSocketManager) handleToolResponse(requestID, tool, command string, result interface{}) {
	m.requestMu.RLock()
	responseChan, exists := m.pendingRequests[requestID]
	m.requestMu.RUnlock()

	if !exists {
		return
	}

	// For text_editor tool
	if tool == "text_editor" && command == "view" {
		if typedChan, ok := responseChan.(chan *textEditorViewOutput); ok {
			// Convert result to JSON bytes then unmarshal directly into target type
			jsonBytes, err := json.Marshal(result)
			if err != nil {
				typedChan <- &textEditorViewOutput{Error: "Failed to marshal response"}
				return
			}

			// Unmarshal directly into the target type
			var output textEditorViewOutput
			if err := json.Unmarshal(jsonBytes, &output); err != nil {
				typedChan <- &textEditorViewOutput{Error: "Failed to parse response"}
				return
			}

			typedChan <- &output
		}
	}
	// Future: Add handling for other tools here
}

// Helper function to generate unique request IDs
func generateRequestID() string {
	return uuid.New().String()
}
