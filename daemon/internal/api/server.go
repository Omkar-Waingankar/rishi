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

	// Streaming chat endpoint (NDJSON)
	r.Post("/chat", s.handleChat)

	// API key management endpoints
	r.Get("/api/key", s.handleGetAPIKey)
	r.Post("/api/key", s.handleSetAPIKey)
	r.Post("/api/key/validate", s.handleValidateAPIKey)

	return r
}
