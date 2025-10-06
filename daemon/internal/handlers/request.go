package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	defaultMaxTokens = 8192
)

// inboundContent defines content types for inbound messages
type inboundContent struct {
	Type       string `json:"type"`                 // "text" | "image"
	Content    string `json:"content,omitempty"`    // for text content
	MediaType  string `json:"mediaType,omitempty"`  // for image content
	DataBase64 string `json:"dataBase64,omitempty"` // for image content
}

// Common request structures
type inboundMessage struct {
	Role    string           `json:"role"`
	Content []inboundContent `json:"content"`
}

type reqBody struct {
	History []inboundMessage `json:"history"`
	Content []inboundContent `json:"content"`
	MaxTok  int              `json:"max_tokens"`
}

// Common request validation
func validateChatRequest(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return false
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// Common request parsing
func parseChatRequest(r *http.Request) (reqBody, error) {
	var in reqBody
	err := json.NewDecoder(r.Body).Decode(&in) // tolerate empty/malformed JSON
	if err != nil {
		return reqBody{}, err
	}
	return in, nil
}

// Common response headers setup
func setupStreamingHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

// Common flusher validation
func validateFlusher(w http.ResponseWriter) (http.Flusher, bool) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return nil, false
	}
	return flusher, true
}

// Common API key validation
func validateAPIKey(w http.ResponseWriter, r *http.Request) (string, bool) {
	apiKey := r.Header.Get("X-Provider-API-Key")
	if apiKey == "" {
		http.Error(w, "missing X-Provider-API-Key header", http.StatusUnauthorized)
		return "", false
	}
	return apiKey, true
}

const maxImageSize = 5 * 1024 * 1024 // 5MB per image

// validateImageContent validates image content blocks
func validateImageContent(content inboundContent) error {
	// Validate media type
	switch content.MediaType {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		// Valid types
	default:
		return fmt.Errorf("unsupported image media type: %s", content.MediaType)
	}

	// Validate and decode base64 data
	if content.DataBase64 == "" {
		return fmt.Errorf("missing image data")
	}

	data, err := base64.StdEncoding.DecodeString(content.DataBase64)
	if err != nil {
		return fmt.Errorf("invalid base64 image data: %v", err)
	}

	// Validate size
	if len(data) > maxImageSize {
		return fmt.Errorf("image too large: %d bytes (max %d bytes)", len(data), maxImageSize)
	}

	return nil
}
