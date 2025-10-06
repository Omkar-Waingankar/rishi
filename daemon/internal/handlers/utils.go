package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	rishiTools "github.com/Omkar-Waingankar/rishi/daemon/internal/tools"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
)

func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T

	schema := reflector.Reflect(v)

	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}

// streamToolCallStart writes a tool call start event to the response stream so we can see the tool call in the frontend
func streamToolCallStart(w http.ResponseWriter, flusher http.Flusher, name string, input interface{}) {
	_ = json.NewEncoder(w).Encode(map[string]any{
		"tool_call": map[string]any{
			"name":   name,
			"input":  input,
			"status": "requesting",
		},
	})
	flusher.Flush()
}

// streamToolCallComplete writes a tool call completion event to the response stream so we can see the tool call in the frontend
func streamToolCallComplete(w http.ResponseWriter, flusher http.Flusher, name string, input interface{}, result interface{}) bool {
	_ = json.NewEncoder(w).Encode(map[string]any{
		"tool_call": map[string]any{
			"name":   name,
			"input":  input,
			"status": "completed",
			"result": result,
		},
	})
	flusher.Flush()

	// Check if result has an error field so we can show the error in the frontend
	switch r := result.(type) {
	case rishiTools.TextEditorViewOutput:
		return r.Error != ""
	case rishiTools.TextEditorStrReplaceOutput:
		return r.Error != ""
	case rishiTools.TextEditorCreateOutput:
		return r.Error != ""
	case rishiTools.TextEditorInsertOutput:
		return r.Error != ""
	case rishiTools.ConsoleExecOutput:
		return r.Error != ""
	case rishiTools.RHelpOutput:
		return r.Error != ""
	}
	return false
}

// formatMessageContentForDebug formats message content for debug logging
// Shows first 10 and last 10 characters of content, handling edge cases
func formatMessageContentForDebug(content inboundContent) string {
	switch content.Type {
	case "text":
		if content.Content == "" {
			return "[empty text content]"
		}

		text := content.Content
		length := len(text)

		if length <= 20 {
			// If text is 20 chars or less, show the whole thing
			return fmt.Sprintf("[%d chars: '%s']", length, text)
		}

		// Show first 10 and last 10 characters
		first := text[:10]
		last := text[length-10:]
		return fmt.Sprintf("[%d chars: '%s...%s']", length, first, last)

	case "image":
		if content.DataBase64 == "" {
			return "[empty image data]"
		}

		// For images, show media type and data length
		dataLength := len(content.DataBase64)
		return fmt.Sprintf("[image: %s, %d chars base64]", content.MediaType, dataLength)

	default:
		return fmt.Sprintf("[unknown type: %s]", content.Type)
	}
}

// formatMessageContentArrayForDebug formats an array of content blocks for debug logging
func formatMessageContentArrayForDebug(contents []inboundContent) string {
	if len(contents) == 0 {
		return "[no content blocks]"
	}

	if len(contents) == 1 {
		return formatMessageContentForDebug(contents[0])
	}

	var parts []string
	for i, content := range contents {
		parts = append(parts, fmt.Sprintf("%d:%s", i, formatMessageContentForDebug(content)))
	}

	return fmt.Sprintf("[%s]", joinStrings(parts, ", "))
}

// joinStrings is a simple helper to join strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
