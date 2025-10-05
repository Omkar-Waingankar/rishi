package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ServerClient hosts HTTP endpoints for the Rishi backend.
type ServerClient struct{}

func NewServerClient() *ServerClient {
	return &ServerClient{}
}

// Routes returns the HTTP handler with all routes registered.
func (s *ServerClient) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(CORS())

	// Health check endpoint
	r.Get("/health", s.handleHealth)

	// Streaming chat endpoints (NDJSON)
	r.Post("/chat/anthropic", s.handleAnthropicChat)
	r.Post("/chat/openai", s.handleOpenAIChat)

	// API key management endpoints
	r.Get("/api/keys", s.handleGetAllAPIKeys)
	r.Post("/api/key/{provider}", s.handleSetAPIKey)
	r.Post("/api/key/{provider}/validate", s.handleValidateAPIKey)

	return r
}
