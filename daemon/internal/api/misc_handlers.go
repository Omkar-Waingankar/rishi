package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
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

// handleGetAllAPIKeys returns the API key status for both providers
func (s *ServerClient) handleGetAllAPIKeys(w http.ResponseWriter, r *http.Request) {
	// Get Anthropic API key
	anthropicKey, anthropicErr := GetAnthropicAPIKey()
	anthropicHasKey := anthropicKey != "" && anthropicErr == nil

	// Get OpenAI API key
	openaiKey, openaiErr := GetOpenAIAPIKey()
	openaiHasKey := openaiKey != "" && openaiErr == nil

	// Log errors but don't fail the request
	if anthropicErr != nil {
		log.Error().Err(anthropicErr).Msg("Failed to get Anthropic API key")
	}
	if openaiErr != nil {
		log.Error().Err(openaiErr).Msg("Failed to get OpenAI API key")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"anthropic": map[string]interface{}{
			"has_key": anthropicHasKey,
			"api_key": anthropicKey,
		},
		"openai": map[string]interface{}{
			"has_key": openaiHasKey,
			"api_key": openaiKey,
		},
	})
}

// handleValidateAPIKey validates an API key against the specified provider's API
func (s *ServerClient) handleValidateAPIKey(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

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

	var isValid bool
	var err error

	switch provider {
	case "anthropic":
		isValid, err = s.validateAnthropicKey(r.Context(), in.APIKey)
	case "openai":
		isValid, err = s.validateOpenAIKey(r.Context(), in.APIKey)
	default:
		http.Error(w, "invalid provider", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Error().Err(err).Msgf("Error validating %s API key", provider)
		isValid = false
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"valid": isValid})
}

// validateAnthropicKey validates an Anthropic API key
func (s *ServerClient) validateAnthropicKey(ctx context.Context, apiKey string) (bool, error) {
	// Basic format validation
	if !strings.HasPrefix(apiKey, "sk-ant-") {
		return false, nil
	}

	// Test the API key with Anthropic API
	testClient := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	_, err := testClient.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5HaikuLatest,
		MaxTokens: int64(validationMaxTokens),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("hi")),
		},
	})

	// Both 200 (success) and 400 (validation error) mean the API key is valid
	// Only authentication errors (401) mean the key is invalid
	return err == nil || !strings.Contains(err.Error(), "401"), nil
}

// validateOpenAIKey validates an OpenAI API key
func (s *ServerClient) validateOpenAIKey(ctx context.Context, apiKey string) (bool, error) {
	// Basic format validation
	if !strings.HasPrefix(apiKey, "sk-") {
		return false, nil
	}

	// Test the API key with OpenAI API
	client := openai.NewClient(apiKey)

	_, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     openai.GPT4oMini,
		MaxTokens: validationMaxTokens,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "hi",
			},
		},
	})

	// Both 200 (success) and 400 (validation error) mean the API key is valid
	// Only authentication errors (401) mean the key is invalid
	return err == nil || !strings.Contains(err.Error(), "401"), nil
}

// handleSetAPIKey saves an API key to the config for the specified provider
func (s *ServerClient) handleSetAPIKey(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

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

	var err error
	switch provider {
	case "anthropic":
		err = SetAnthropicAPIKey(in.APIKey)
	case "openai":
		err = SetOpenAIAPIKey(in.APIKey)
	default:
		http.Error(w, "invalid provider", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Error().Err(err).Msgf("Failed to save %s API key", provider)
		http.Error(w, "failed to save API key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
