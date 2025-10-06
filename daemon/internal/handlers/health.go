package handlers

import (
	"encoding/json"
	"net/http"
)

const (
	validationMaxTokens = 1
)

// HandleHealth returns a simple health check response
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "rishi-daemon",
	})
}
