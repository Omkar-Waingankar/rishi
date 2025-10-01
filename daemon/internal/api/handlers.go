package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/rs/zerolog/log"
)

// handleHealth returns a simple health check response
func (s *ServerClient) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "rishi-daemon",
	})
}

// handleGetAPIKey returns the API key from the config
func (s *ServerClient) handleGetAPIKey(w http.ResponseWriter, r *http.Request) {
	apiKey, err := GetAPIKey()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get API key")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"has_key": false,
			"api_key": "",
		})
		return
	}

	hasKey := apiKey != ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"has_key": hasKey,
		"api_key": apiKey,
	})
}

// handleValidateAPIKey validates an API key against the Anthropic API
func (s *ServerClient) handleValidateAPIKey(w http.ResponseWriter, r *http.Request) {
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

	// Basic format validation
	if !strings.HasPrefix(in.APIKey, "sk-ant-") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{
			"valid": false,
		})
		return
	}

	if len(in.APIKey) < 20 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{
			"valid": false,
		})
		return
	}

	// Test the API key with Anthropic API
	testClient := anthropic.NewClient(
		option.WithAPIKey(in.APIKey),
	)

	_, err := testClient.Messages.New(r.Context(), anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5HaikuLatest,
		MaxTokens: int64(1),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("hi")),
		},
	})

	// Both 200 (success) and 400 (validation error) mean the API key is valid
	// Only authentication errors (401) mean the key is invalid
	isValid := err == nil || !strings.Contains(err.Error(), "401")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{
		"valid": isValid,
	})
}

// handleSetAPIKey saves an API key to the config
func (s *ServerClient) handleSetAPIKey(w http.ResponseWriter, r *http.Request) {
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

	if err := SetAPIKey(in.APIKey); err != nil {
		log.Error().Err(err).Msg("Failed to save API key")
		http.Error(w, "failed to save API key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{
		"success": true,
	})
}

// handleChat proxies a streaming request with history to Anthropic and emits NDJSON lines
// of the form {"text": "..."} and a final {"is_final": true}.
func (s *ServerClient) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get API key from header
	apiKey := r.Header.Get("X-Anthropic-API-Key")
	if apiKey == "" {
		http.Error(w, "missing X-Anthropic-API-Key header", http.StatusUnauthorized)
		return
	}

	// Create Anthropic client for this request
	anthropicClient := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	type inboundMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type reqBody struct {
		History []inboundMessage `json:"history"`
		Message string           `json:"message"`
		Model   string           `json:"model"`
		MaxTok  int              `json:"max_tokens"`
	}
	var in reqBody
	_ = json.NewDecoder(r.Body).Decode(&in) // tolerate empty/malformed JSON

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Convert history into []anthropic.MessageParam, include system prompt, then append latest user message
	var msgs []anthropic.MessageParam
	// Prepend system prompt as a message to keep behavior similar
	msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(RISHI_SYSTEM_PROMPT)))
	for i, m := range in.History {
		switch m.Role {
		case "user":
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		case "assistant":
			msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		default:
			// ignore
		}
		log.Info().Msgf("history message %d: role %s, content %s", i, m.Role, m.Content[:min(100, len(m.Content))])
	}

	if in.Message != "" {
		msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(in.Message)))
	}
	log.Info().Msgf("new user message: %s", in.Message[:min(100, len(in.Message))])

	// Start streaming with the official Anthropic SDK
	model := anthropic.ModelClaudeSonnet4_20250514

	// Check for model in request body first, then X-Model header
	selectedModel := in.Model
	if selectedModel == "" {
		selectedModel = r.Header.Get("X-Model")
	}

	if selectedModel != "" {
		// Map model names from frontend to Anthropic SDK models
		switch selectedModel {
		case "claude-3.7-sonnet":
			model = anthropic.ModelClaude3_7SonnetLatest
			log.Info().Msgf("Using Claude 3.7 Sonnet model")
		case "claude-4-sonnet":
			model = anthropic.ModelClaudeSonnet4_20250514
			log.Info().Msgf("Using Claude 4 Sonnet model")
		default:
			// If unknown model, log and use default
			log.Warn().Msgf("Unknown model requested: %s, using default Claude 4 Sonnet", selectedModel)
		}
	} else {
		log.Info().Msgf("No model specified, using default Claude 4 Sonnet")
	}

	maxTokens := in.MaxTok
	if maxTokens == 0 {
		maxTokens = 8192
	}

	tools := []anthropic.ToolUnionParam{}
	if selectedModel == "claude-4-sonnet" {
		tools = append(tools, anthropic.ToolUnionParam{OfTextEditor20250728: &anthropic.ToolTextEditor20250728Param{}})
	} else if selectedModel == "claude-3.7-sonnet" {
		tools = append(tools, anthropic.ToolUnionParam{OfTextEditor20250124: &anthropic.ToolTextEditor20250124Param{}})
	}

	for {
		stream := anthropicClient.Messages.NewStreaming(r.Context(), anthropic.MessageNewParams{
			Model:       model,
			MaxTokens:   int64(maxTokens),
			Messages:    msgs,
			Tools:       tools,
			Temperature: anthropic.Opt(0.1),
		})

		message := anthropic.Message{}
		for stream.Next() {
			event := stream.Current()
			if err := message.Accumulate(event); err != nil {
				log.Error().Err(err).Msg("message accumulation error")

				// Parse accumulation errors
				errorMsg := err.Error()
				var friendlyMsg string

				if strings.Contains(errorMsg, "overloaded_error") || strings.Contains(errorMsg, "Overloaded") {
					friendlyMsg = "Claude is currently experiencing high demand. Please try again in a few moments."
				} else {
					friendlyMsg = fmt.Sprintf("Error processing response: %v", err)
				}

				_ = json.NewEncoder(w).Encode(map[string]any{"error": friendlyMsg})
				flusher.Flush()
				break
			}

			switch eventVariant := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				switch deltaVariant := eventVariant.Delta.AsAny().(type) {
				case anthropic.TextDelta:
					_ = json.NewEncoder(w).Encode(map[string]any{"text": deltaVariant.Text})
					flusher.Flush()
				}
			}
		}

		// Check for streaming errors
		if err := stream.Err(); err != nil {
			log.Error().Err(err).Msg("streaming error occurred")

			// Parse common Anthropic API errors to provide user-friendly messages
			errorMsg := err.Error()
			var friendlyMsg string

			if strings.Contains(errorMsg, "overloaded_error") || strings.Contains(errorMsg, "Overloaded") {
				friendlyMsg = "Claude is currently experiencing high demand. Please try again in a few moments."
			} else {
				friendlyMsg = fmt.Sprintf("Claude encountered an error: %v", err)
			}

			_ = json.NewEncoder(w).Encode(map[string]any{"error": friendlyMsg})
			flusher.Flush()
			return
		}

		toolResults := []anthropic.ContentBlockParamUnion{}
		for _, block := range message.Content {
			switch variant := block.AsAny().(type) {
			case anthropic.TextBlock:
				msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(variant.Text)))
			case anthropic.ToolUseBlock:
				log.Info().Msgf("tool use: %s, input: %s", block.Name, variant.JSON.Input.Raw())

				var response interface{}
				switch block.Name {
				case "str_replace_based_edit_tool":
					var input textEditorInput
					if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input); err != nil {
						errMsg := fmt.Sprintf("Failed to parse text editor input: %s, error: %v", variant.JSON.Input.Raw(), err)
						log.Error().Err(err).Msgf(errMsg)
						response = textEditorViewOutput{
							Error: errMsg,
						}
						break
					}

					// Validate required fields
					if input.Command == "" {
						errMsg := "Error: Missing required 'command' field. The text editor tool requires a 'command' parameter. Available commands: 'view' (to read files/directories). Example: {\"command\": \"view\", \"path\": \"filename.txt\"}"
						log.Error().Msg(errMsg)
						response = textEditorViewOutput{
							Error: errMsg,
						}
						break
					}

					switch input.Command {
					case ViewCommand:
						// Stream tool call start event to frontend
						viewInput := textEditorViewInput{
							Path:      input.Path,
							ViewRange: input.ViewRange,
						}
						_ = json.NewEncoder(w).Encode(map[string]any{
							"tool_call": map[string]any{
								"name":   input.Command,
								"input":  viewInput,
								"status": "requesting",
							},
						})
						flusher.Flush()

						// Execute text editor view command
						response = textEditorView(viewInput)
					case StrReplaceCommand:
						// Stream tool call start event to frontend
						strReplaceInput := textEditorStrReplaceInput{
							Path:   input.Path,
							OldStr: input.OldStr,
							NewStr: input.NewStr,
						}
						_ = json.NewEncoder(w).Encode(map[string]any{
							"tool_call": map[string]any{
								"name":   input.Command,
								"input":  strReplaceInput,
								"status": "requesting",
							},
						})
						flusher.Flush()

						// Execute text editor str_replace command
						response = textEditorStrReplace(strReplaceInput)
					case CreateCommand:
						// Stream tool call start event to frontend
						createInput := textEditorCreateInput{
							Path:     input.Path,
							FileText: input.FileText,
						}
						_ = json.NewEncoder(w).Encode(map[string]any{
							"tool_call": map[string]any{
								"name":   input.Command,
								"input":  createInput,
								"status": "requesting",
							},
						})
						flusher.Flush()

						// Execute text editor create command
						response = textEditorCreate(createInput)
					case InsertCommand:
						// Stream tool call start event to frontend
						// Handle both field names - docs say new_str but API sends insert_text
						insertText := input.NewStr
						if insertText == "" {
							insertText = input.InsertText
						}
						insertInput := textEditorInsertInput{
							Path:       input.Path,
							InsertLine: input.InsertLine,
							NewStr:     insertText,
						}
						_ = json.NewEncoder(w).Encode(map[string]any{
							"tool_call": map[string]any{
								"name":   input.Command,
								"input":  insertInput,
								"status": "requesting",
							},
						})
						flusher.Flush()

						// Execute text editor insert command
						response = textEditorInsert(insertInput)
					}
				}

				b, err := json.Marshal(response)
				if err != nil {
					http.Error(w, "error parsing tool result", http.StatusInternalServerError)
					return
				}

				log.Info().Msgf("tool call completed: %s, result length: %d, result: %s", block.Name, len(string(b)), string(b)[:min(100, len(string(b)))])

				var isError bool

				// Stream tool call completion event to frontend
				switch block.Name {
				case "str_replace_based_edit_tool":
					var input textEditorInput
					if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input); err != nil {
						errMsg := fmt.Sprintf("Failed to parse read_file input: %s, error: %v", variant.JSON.Input.Raw(), err)
						log.Error().Err(err).Msgf(errMsg)
					}

					switch response := response.(type) {
					case textEditorViewOutput:
						_ = json.NewEncoder(w).Encode(map[string]any{
							"tool_call": map[string]any{
								"name":   ViewCommand,
								"input":  input,
								"status": "completed",
								"result": response,
							},
						})
						flusher.Flush()

						if response.Error != "" {
							isError = true
						}
					case textEditorStrReplaceOutput:
						_ = json.NewEncoder(w).Encode(map[string]any{
							"tool_call": map[string]any{
								"name":   StrReplaceCommand,
								"input":  input,
								"status": "completed",
								"result": response,
							},
						})
						flusher.Flush()

						if response.Error != "" {
							isError = true
						}
					case textEditorCreateOutput:
						_ = json.NewEncoder(w).Encode(map[string]any{
							"tool_call": map[string]any{
								"name":   CreateCommand,
								"input":  input,
								"status": "completed",
								"result": response,
							},
						})
						flusher.Flush()

						if response.Error != "" {
							isError = true
						}
					case textEditorInsertOutput:
						_ = json.NewEncoder(w).Encode(map[string]any{
							"tool_call": map[string]any{
								"name":   InsertCommand,
								"input":  input,
								"status": "completed",
								"result": response,
							},
						})
						flusher.Flush()

						if response.Error != "" {
							isError = true
						}
					}
				}

				msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewToolUseBlock(block.ID, json.RawMessage(variant.JSON.Input.Raw()), block.Name)))

				toolResults = append(toolResults, anthropic.NewToolResultBlock(block.ID, string(b), isError))
				msgs = append(msgs, anthropic.NewUserMessage(toolResults...))
			}
		}

		if len(toolResults) == 0 {
			// If no tool results, we're done streaming
			if stream.Err() != nil {
				_ = json.NewEncoder(w).Encode(map[string]any{"error": stream.Err().Error()})
				flusher.Flush()
				return
			}

			_ = json.NewEncoder(w).Encode(map[string]any{"is_final": true})
			flusher.Flush()

			break
		}
	}

}
