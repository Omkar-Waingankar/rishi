package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/rs/zerolog/log"
)

const (
	defaultMaxTokens = 8192
)

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
		Role    string           `json:"role"`
		Content []inboundContent `json:"content"`
	}

	type reqBody struct {
		History []inboundMessage `json:"history"`
		Content []inboundContent `json:"content"` // Changed from Message string
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
			// Convert content blocks for user messages
			contentBlocks, err := convertToAnthropicContent(m.Content)
			if err != nil {
				log.Error().Err(err).Msgf("Error converting user history content")
				http.Error(w, fmt.Sprintf("Invalid user message content: %v", err), http.StatusBadRequest)
				return
			}
			if len(contentBlocks) > 0 {
				msgs = append(msgs, anthropic.NewUserMessage(contentBlocks...))
			}
		case "assistant":
			// For assistant messages in history, we only have text content
			// Extract text content from the content array
			var textContent string
			for _, content := range m.Content {
				if content.Type == "text" {
					textContent += content.Content
				}
			}
			if textContent != "" {
				msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(textContent)))
			}
		default:
			// ignore
		}
		log.Info().Msgf("history message %d: role %s, %d content blocks", i, m.Role, len(m.Content))
	}

	// Handle the new user message content
	if len(in.Content) > 0 {
		contentBlocks, err := convertToAnthropicContent(in.Content)
		if err != nil {
			log.Error().Err(err).Msgf("Error converting user message content")
			http.Error(w, fmt.Sprintf("Invalid message content: %v", err), http.StatusBadRequest)
			return
		}
		if len(contentBlocks) > 0 {
			msgs = append(msgs, anthropic.NewUserMessage(contentBlocks...))
		}
	}
	log.Info().Msgf("new user message: %d content blocks", len(in.Content))

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
		maxTokens = defaultMaxTokens
	}

	tools := []anthropic.ToolUnionParam{}
	if selectedModel == "claude-4-sonnet" {
		tools = append(tools, anthropic.ToolUnionParam{OfTextEditor20250728: &anthropic.ToolTextEditor20250728Param{}})
	} else if selectedModel == "claude-3.7-sonnet" {
		tools = append(tools, anthropic.ToolUnionParam{OfTextEditor20250124: &anthropic.ToolTextEditor20250124Param{}})
	}

	// Add custom console tools
	consoleExecTool := anthropic.ToolParam{
		Name:        "console_exec",
		Description: anthropic.String("Executes R code in the user's R console. The code will be sent to the console and executed immediately."),
		InputSchema: GenerateSchema[consoleExecInput](),
	}

	tools = append(tools, anthropic.ToolUnionParam{OfTool: &consoleExecTool})

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
				_ = json.NewEncoder(w).Encode(map[string]any{"error": parseAnthropicError(err)})
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
			_ = json.NewEncoder(w).Encode(map[string]any{"error": parseAnthropicError(err)})
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
				case "console_exec":
					var input consoleExecInput
					if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input); err != nil {
						errMsg := fmt.Sprintf("Failed to parse console exec input: %s, error: %v", variant.JSON.Input.Raw(), err)
						log.Error().Err(err).Msgf(errMsg)
						response = consoleExecOutput{
							Error: errMsg,
						}
						break
					}

					streamToolCallStart(w, flusher, "console_exec", input)
					response = consoleExec(input)

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
						viewInput := textEditorViewInput{
							Path:      input.Path,
							ViewRange: input.ViewRange,
						}
						streamToolCallStart(w, flusher, string(input.Command), viewInput)
						response = textEditorView(viewInput)
					case StrReplaceCommand:
						strReplaceInput := textEditorStrReplaceInput{
							Path:   input.Path,
							OldStr: input.OldStr,
							NewStr: input.NewStr,
						}
						streamToolCallStart(w, flusher, string(input.Command), strReplaceInput)
						response = textEditorStrReplace(strReplaceInput)
					case CreateCommand:
						createInput := textEditorCreateInput{
							Path:     input.Path,
							FileText: input.FileText,
						}
						streamToolCallStart(w, flusher, string(input.Command), createInput)
						response = textEditorCreate(createInput)
					case InsertCommand:
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
						streamToolCallStart(w, flusher, string(input.Command), insertInput)
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
				case "console_exec":
					var input consoleExecInput
					if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input); err != nil {
						log.Error().Err(err).Msgf("Failed to parse console exec input for completion event")
					}

					isError = streamToolCallComplete(w, flusher, "console_exec", input, response)
				case "str_replace_based_edit_tool":
					var input textEditorInput
					if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input); err != nil {
						errMsg := fmt.Sprintf("Failed to parse text editor input: %s, error: %v", variant.JSON.Input.Raw(), err)
						log.Error().Err(err).Msgf(errMsg)
					}

					var commandName string
					switch response.(type) {
					case textEditorViewOutput:
						commandName = string(ViewCommand)
					case textEditorStrReplaceOutput:
						commandName = string(StrReplaceCommand)
					case textEditorCreateOutput:
						commandName = string(CreateCommand)
					case textEditorInsertOutput:
						commandName = string(InsertCommand)
					}

					isError = streamToolCallComplete(w, flusher, commandName, input, response)
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
