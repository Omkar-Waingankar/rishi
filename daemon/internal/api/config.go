package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog/log"
)

// Config represents the Rishi configuration structure
type Config struct {
	AnthropicAPIKey string `json:"anthropic_api_key,omitempty"`
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

// GetAPIKey retrieves the ANTHROPIC_API_KEY from the config file
func GetAPIKey() (string, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}
	return config.AnthropicAPIKey, nil
}

// SetAPIKey saves the ANTHROPIC_API_KEY to the config file
func SetAPIKey(apiKey string) error {
	config, err := LoadConfig()
	if err != nil {
		// If we can't load config, start with empty config
		config = &Config{}
	}

	config.AnthropicAPIKey = apiKey
	return SaveConfig(config)
}
