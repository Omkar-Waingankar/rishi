package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
	"github.com/rs/zerolog/log"
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
	case textEditorViewOutput:
		return r.Error != ""
	case textEditorStrReplaceOutput:
		return r.Error != ""
	case textEditorCreateOutput:
		return r.Error != ""
	case textEditorInsertOutput:
		return r.Error != ""
	}
	return false
}

// inboundContent defines content types for inbound messages
type inboundContent struct {
	Type       string `json:"type"`                 // "text" | "image"
	Content    string `json:"content,omitempty"`    // for text content
	MediaType  string `json:"mediaType,omitempty"`  // for image content
	DataBase64 string `json:"dataBase64,omitempty"` // for image content
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

// convertToAnthropicContent converts inbound content to Anthropic content blocks
func convertToAnthropicContent(contents []inboundContent) ([]anthropic.ContentBlockParamUnion, error) {
	var blocks []anthropic.ContentBlockParamUnion

	for _, content := range contents {
		switch content.Type {
		case "text":
			if content.Content != "" {
				blocks = append(blocks, anthropic.NewTextBlock(content.Content))
			}
		case "image":
			if err := validateImageContent(content); err != nil {
				return nil, fmt.Errorf("invalid image content: %v", err)
			}

			blocks = append(blocks, anthropic.NewImageBlockBase64(content.MediaType, content.DataBase64))
		default:
			log.Warn().Msgf("Unknown content type: %s", content.Type)
		}
	}

	return blocks, nil
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

// convertToOpenaiContent converts a content array to a string for OpenAI
func convertToOpenaiContent(contents []inboundContent) string {
	var textContent string

	for _, content := range contents {
		switch content.Type {
		case "text":
			if content.Content != "" {
				textContent += content.Content
			}
		case "image":
			// For images, we can't easily convert to text, so we'll skip them for now
			// OpenAI doesn't support image content blocks in the same way as Anthropic
			log.Warn().Msgf("Skipping image content in OpenAI message conversion")
		default:
			log.Warn().Msgf("Unknown content type in OpenAI message conversion: %s", content.Type)
		}
	}

	return textContent
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
