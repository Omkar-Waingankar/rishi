package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/rs/zerolog/log"
)

const (
	validationMaxTokens = 1
)

// handleHealth returns a simple health check response
func (s *ServerClient) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "rishi-daemon",
	})
}

// handleGetAPIKey returns the API key from the config
func (s *ServerClient) handleGetAPIKey(w http.ResponseWriter, r *http.Request) {
	apiKey, err := GetAPIKey()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get API key")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"has_key": false,
			"api_key": "",
		})
		return
	}

	hasKey := apiKey != ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"has_key": hasKey,
		"api_key": apiKey,
	})
}

// handleValidateAPIKey validates an API key against the Anthropic API
func (s *ServerClient) handleValidateAPIKey(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		APIKey string `json:"api_key"`
	}
	var in reqBody
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if in.APIKey == "" {
		http.Error(w, "missing api_key parameter", http.StatusBadRequest)
		return
	}

	// Basic format validation
	if !strings.HasPrefix(in.APIKey, "sk-ant-") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"valid": false})
		return
	}

	if len(in.APIKey) < 20 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"valid": false})
		return
	}

	// Test the API key with Anthropic API
	testClient := anthropic.NewClient(
		option.WithAPIKey(in.APIKey),
	)

	_, err := testClient.Messages.New(r.Context(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5HaikuLatest,
		MaxTokens: int64(validationMaxTokens),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("hi")),
		},
	})

	// Both 200 (success) and 400 (validation error) mean the API key is valid
	// Only authentication errors (401) mean the key is invalid
	isValid := err == nil || !strings.Contains(err.Error(), "401")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"valid": isValid})
}

// handleSetAPIKey saves an API key to the config
func (s *ServerClient) handleSetAPIKey(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		APIKey string `json:"api_key"`
	}
	var in reqBody
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if in.APIKey == "" {
		http.Error(w, "missing api_key parameter", http.StatusBadRequest)
		return
	}

	if err := SetAPIKey(in.APIKey); err != nil {
		log.Error().Err(err).Msg("Failed to save API key")
		http.Error(w, "failed to save API key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
