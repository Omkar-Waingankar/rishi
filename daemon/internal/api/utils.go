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
