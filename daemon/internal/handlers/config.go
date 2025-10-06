package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
)

// Config represents the Rishi configuration structure
type Config struct {
	AnthropicAPIKey string `json:"anthropic_api_key,omitempty"`
	OpenAIAPIKey    string `json:"openai_api_key,omitempty"`
}

// handleGetAllAPIKeys returns the API key status for both providers
func HandleGetAllAPIKeys(w http.ResponseWriter, r *http.Request) {
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
func HandleValidateAPIKey(w http.ResponseWriter, r *http.Request) {
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
		isValid, err = validateAnthropicKey(r.Context(), in.APIKey)
	case "openai":
		isValid, err = validateOpenAIKey(r.Context(), in.APIKey)
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
func validateAnthropicKey(ctx context.Context, apiKey string) (bool, error) {
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
func validateOpenAIKey(ctx context.Context, apiKey string) (bool, error) {
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
func HandleSetAPIKey(w http.ResponseWriter, r *http.Request) {
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

// getConfigDir returns the platform-appropriate config directory path for Rishi
func getConfigDir() (string, error) {
	var configDir string

	if runtime.GOOS == "windows" {
		// Windows: use APPDATA or LOCALAPPDATA
		configBase := os.Getenv("APPDATA")
		if configBase == "" {
			configBase = os.Getenv("LOCALAPPDATA")
		}
		if configBase == "" {
			configBase = os.Getenv("HOME")
		}
		configDir = filepath.Join(configBase, "rishi")
	} else {
		// Unix/Mac: use XDG_CONFIG_HOME or ~/.config
		configBase := os.Getenv("XDG_CONFIG_HOME")
		if configBase == "" {
			home := os.Getenv("HOME")
			if home == "" {
				return "", fmt.Errorf("HOME environment variable not set")
			}
			configBase = filepath.Join(home, ".config")
		}
		configDir = filepath.Join(configBase, "rishi")
	}

	return configDir, nil
}

// getConfigPath returns the full path to the config.json file
func getConfigPath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

// LoadConfig reads the config file and returns a Config struct
func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig writes the config to the config file
func SaveConfig(config *Config) error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Marshal config to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to temp file first for atomic write
	tempFile, err := os.CreateTemp(configDir, "config.json.*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath) // Clean up if we fail

	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write config: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Set restrictive permissions (Unix only, no-op on Windows)
	if err := os.Chmod(tempPath, 0600); err != nil {
		log.Warn().Err(err).Msg("Failed to set config file permissions")
	}

	// Move temp file to final location (atomic on most systems)
	if err := os.Rename(tempPath, configPath); err != nil {
		return fmt.Errorf("failed to move config file: %w", err)
	}

	return nil
}

// GetAnthropicAPIKey retrieves the ANTHROPIC_API_KEY from the config file
func GetAnthropicAPIKey() (string, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}
	return config.AnthropicAPIKey, nil
}

// SetAnthropicAPIKey saves the ANTHROPIC_API_KEY to the config file
func SetAnthropicAPIKey(apiKey string) error {
	config, err := LoadConfig()
	if err != nil {
		// If we can't load config, start with empty config
		config = &Config{}
	}

	config.AnthropicAPIKey = apiKey
	return SaveConfig(config)
}

// GetOpenAIAPIKey retrieves the OPENAI_API_KEY from the config file
func GetOpenAIAPIKey() (string, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}
	return config.OpenAIAPIKey, nil
}

func SetOpenAIAPIKey(apiKey string) error {
	config, err := LoadConfig()
	if err != nil {
		// If we can't load config, start with empty config
		config = &Config{}
	}
	config.OpenAIAPIKey = apiKey
	return SaveConfig(config)
}
