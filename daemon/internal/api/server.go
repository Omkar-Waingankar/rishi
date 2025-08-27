package api

import (
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/go-chi/chi/v5"
)

// ServerClient hosts HTTP endpoints for the R-Agent backend.
type ServerClient struct {
	anthropicClient anthropic.Client
	rstudioURL      string
}

func NewServerClient(anthropicClient anthropic.Client, rstudioURL string) *ServerClient {
	return &ServerClient{
		anthropicClient: anthropicClient,
		rstudioURL:      rstudioURL,
	}
}

// Routes returns the HTTP handler with all routes registered.
func (s *ServerClient) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(CORS())

	// Streaming chat endpoint (NDJSON)
	r.Post("/chat", s.handleChat)

	return r
}
