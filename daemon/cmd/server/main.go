package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rstudio-ai-server/ragent/internal/api"
)

type Config struct {
	AnthropicAPIKey string `envconfig:"ANTHROPIC_API_KEY" required:"true"`
	HTTPPort        string `envconfig:"HTTP_PORT" default:"8090"`
	RStudioURL      string `envconfig:"RSTUDIO_URL" default:"http://127.0.0.1:8787"`
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

	anthropicClient := anthropic.NewClient(
		option.WithAPIKey(cfg.AnthropicAPIKey),
	)

	// Build and start HTTP API server
	srv := api.NewServerClient(anthropicClient, cfg.RStudioURL)
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
