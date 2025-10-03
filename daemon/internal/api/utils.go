package api

import (
	"encoding/json"
	"net/http"

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
