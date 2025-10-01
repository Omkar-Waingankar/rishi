package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/halliday/rishi/daemon/internal/api"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Config struct {
	HTTPPort string `envconfig:"HTTP_PORT" default:"8080"`
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(os.Stdout)

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Info().Msg("No .env file found, using system environment variables")
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to process environment configuration")
	}

	// Build and start HTTP API server
	srv := api.NewServerClient()
	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Info().Str("http_port", cfg.HTTPPort).Msg("Starting HTTP server")
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("HTTP server failed to start")
	}
}
