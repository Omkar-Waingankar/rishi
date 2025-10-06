package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	handlers "github.com/Omkar-Waingankar/rishi/daemon/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	defaultHTTPPort = "8080"
)

// setupRoutes returns the HTTP handler with all routes registered.
func routes() http.Handler {
	r := chi.NewRouter()
	r.Use(cors())

	// Health check endpoint
	r.Get("/health", handlers.HandleHealth)

	// Streaming chat endpoints (NDJSON)
	r.Post("/chat/anthropic", handlers.HandleAnthropicChat)
	r.Post("/chat/openai", handlers.HandleOpenAIChat)

	// API key management endpoints
	r.Get("/api/keys", handlers.HandleGetAllAPIKeys)
	r.Post("/api/key/{provider}", handlers.HandleSetAPIKey)
	r.Post("/api/key/{provider}/validate", handlers.HandleValidateAPIKey)

	return r
}

// cors returns a middleware that adds permissive CORS headers and handles OPTIONS preflight.
func cors() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,X-Provider-API-Key,X-Model")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(os.Stdout)

	// Build and start HTTP API server
	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%s", defaultHTTPPort),
		Handler:           routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Info().Str("http_port", defaultHTTPPort).Msg("Starting HTTP server")
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("HTTP server failed to start")
	}
}
