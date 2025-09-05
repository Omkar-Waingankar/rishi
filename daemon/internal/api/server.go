package api

import (
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/go-chi/chi/v5"
)

// ServerClient hosts HTTP endpoints for the Rishi backend.
type ServerClient struct {
	anthropicClient anthropic.Client
	toolRPCToken    string
	wsManager       *WebSocketManager
}

func NewServerClient(anthropicClient anthropic.Client, toolRPCToken string) *ServerClient {
	return &ServerClient{
		anthropicClient: anthropicClient,
		toolRPCToken:    toolRPCToken,
		wsManager:       NewWebSocketManager(),
	}
}

// Routes returns the HTTP handler with all routes registered.
func (s *ServerClient) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(CORS())

	// Streaming chat endpoint (NDJSON)
	r.Post("/chat", s.handleChat)

	// WebSocket endpoint for tool communication
	r.Get("/ws/tools", s.HandleWebSocket)

	return r
}
