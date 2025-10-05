package api

import (
	"encoding/json"
	"net/http"
)

const (
	defaultMaxTokens = 8192
)

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
