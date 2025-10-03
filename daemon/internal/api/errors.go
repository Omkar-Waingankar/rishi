package api

import (
	"fmt"
	"strings"
)

// parseAnthropicError converts Anthropic API errors into user-friendly messages
func parseAnthropicError(err error) string {
	errorMsg := err.Error()
	if strings.Contains(errorMsg, "overloaded_error") || strings.Contains(errorMsg, "Overloaded") {
		return "Claude is currently experiencing high demand. Please try again in a few moments."
	}
	return fmt.Sprintf("Claude encountered an error: %v", err)
}
