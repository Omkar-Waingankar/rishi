package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	rishiTools "github.com/Omkar-Waingankar/rishi/daemon/internal/tools"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/rs/zerolog/log"
)

// handleAnthropicChat proxies a streaming request with history to Anthropic and emits NDJSON lines
// of the form {"text": "..."} and a final {"is_final": true}.
func HandleAnthropicChat(w http.ResponseWriter, r *http.Request) {
	if !validateChatRequest(w, r) {
		return
	}

	// Get API key from header
	apiKey, ok := validateAPIKey(w, r)
	if !ok {
		return
	}

	// Create Anthropic client for this request
	anthropicClient := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	in, err := parseChatRequest(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse chat request")
		http.Error(w, "Failed to parse chat request", http.StatusBadRequest)
		return
	}
	setupStreamingHeaders(w)

	flusher, ok := validateFlusher(w)
	if !ok {
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
		log.Info().Msgf("history message %d: role %s, %d content blocks, content: %s", i, m.Role, len(m.Content), formatMessageContentArrayForDebug(m.Content))
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
	log.Info().Msgf("new user message: %d content blocks, content: %s", len(in.Content), formatMessageContentArrayForDebug(in.Content))

	// Start streaming with the official Anthropic SDK
	model := anthropic.ModelClaudeSonnet4_20250514

	// Get model from X-Model header
	selectedModel := r.Header.Get("X-Model")

	if selectedModel != "" {
		// Map model names from frontend to Anthropic SDK models
		switch selectedModel {
		case "claude-3.7-sonnet":
			model = anthropic.ModelClaude3_7SonnetLatest
			log.Info().Msgf("Using Claude 3.7 Sonnet model")
		case "claude-4-sonnet":
			model = anthropic.ModelClaudeSonnet4_0
			log.Info().Msgf("Using Claude 4 Sonnet model")
		case "claude-4.5-sonnet":
			model = anthropic.ModelClaudeSonnet4_5
			log.Info().Msgf("Using Claude 4.5 Sonnet model")
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
		InputSchema: GenerateSchema[rishiTools.ConsoleExecInput](),
	}

	tools = append(tools, anthropic.ToolUnionParam{OfTool: &consoleExecTool})

	// Add R help tool
	rHelpTool := anthropic.ToolParam{
		Name:        "r_help",
		Description: anthropic.String("Retrieves R package documentation. Use this to look up package-level documentation or specific function/topic documentation when writing R code or answering questions about R packages and their functionality.\n\nParameters:\n- package (required): The name of the R package to get documentation for (e.g., 'dplyr', 'ggplot2', 'rstudioapi').\n- topic (optional): The specific function, method, or topic within the package to get detailed documentation for (e.g., 'addTheme', 'mutate').\n\nExamples:\n- Package-level help: {\"package\": \"ggplot2\"}\n- Topic-specific help: {\"package\": \"rstudioapi\", \"topic\": \"addTheme\"} - equivalent to help(\"addTheme\", package=\"rstudioapi\")"),
		InputSchema: GenerateSchema[rishiTools.RHelpInput](),
	}

	tools = append(tools, anthropic.ToolUnionParam{OfTool: &rHelpTool})

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
				case "r_help":
					var input rishiTools.RHelpInput
					if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input); err != nil {
						errMsg := fmt.Sprintf("Failed to parse r_help input: %s, error: %v", variant.JSON.Input.Raw(), err)
						log.Error().Err(err).Msgf(errMsg)
						response = rishiTools.RHelpOutput{
							Error: errMsg,
						}
						break
					}

					streamToolCallStart(w, flusher, "r_help", input)
					response = rishiTools.RHelp(input)

				case "console_exec":
					var input rishiTools.ConsoleExecInput
					if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input); err != nil {
						errMsg := fmt.Sprintf("Failed to parse console exec input: %s, error: %v", variant.JSON.Input.Raw(), err)
						log.Error().Err(err).Msgf(errMsg)
						response = rishiTools.ConsoleExecOutput{
							Error: errMsg,
						}
						break
					}

					streamToolCallStart(w, flusher, "console_exec", input)
					response = rishiTools.ConsoleExec(input)

				case "str_replace_based_edit_tool":
					var input rishiTools.TextEditorInput
					if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input); err != nil {
						errMsg := fmt.Sprintf("Failed to parse text editor input: %s, error: %v", variant.JSON.Input.Raw(), err)
						log.Error().Err(err).Msgf(errMsg)
						response = rishiTools.TextEditorViewOutput{
							Error: errMsg,
						}
						break
					}

					// Validate required fields
					if input.Command == "" {
						errMsg := "Error: Missing required 'command' field. The text editor tool requires a 'command' parameter. Available commands: 'view' (to read files/directories). Example: {\"command\": \"view\", \"path\": \"filename.txt\"}"
						log.Error().Msg(errMsg)
						response = rishiTools.TextEditorViewOutput{
							Error: errMsg,
						}
						break
					}

					switch input.Command {
					case rishiTools.ViewCommand:
						viewInput := rishiTools.TextEditorViewInput{
							Path:      input.Path,
							ViewRange: input.ViewRange,
						}
						streamToolCallStart(w, flusher, string(input.Command), viewInput)
						response = rishiTools.TextEditorView(viewInput)
					case rishiTools.StrReplaceCommand:
						strReplaceInput := rishiTools.TextEditorStrReplaceInput{
							Path:   input.Path,
							OldStr: input.OldStr,
							NewStr: input.NewStr,
						}
						streamToolCallStart(w, flusher, string(input.Command), strReplaceInput)
						response = rishiTools.TextEditorStrReplace(strReplaceInput)
					case rishiTools.CreateCommand:
						createInput := rishiTools.TextEditorCreateInput{
							Path:     input.Path,
							FileText: input.FileText,
						}
						streamToolCallStart(w, flusher, string(input.Command), createInput)
						response = rishiTools.TextEditorCreate(createInput)
					case rishiTools.InsertCommand:
						// Handle both field names - docs say new_str but API sends insert_text
						insertText := input.NewStr
						if insertText == "" {
							insertText = input.InsertText
						}
						insertInput := rishiTools.TextEditorInsertInput{
							Path:       input.Path,
							InsertLine: input.InsertLine,
							NewStr:     insertText,
						}
						streamToolCallStart(w, flusher, string(input.Command), insertInput)
						response = rishiTools.TextEditorInsert(insertInput)
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
				case "r_help":
					var input rishiTools.RHelpInput
					if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input); err != nil {
						log.Error().Err(err).Msgf("Failed to parse r_help input for completion event")
					}

					isError = streamToolCallComplete(w, flusher, "r_help", input, response)
				case "console_exec":
					var input rishiTools.ConsoleExecInput
					if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input); err != nil {
						log.Error().Err(err).Msgf("Failed to parse console exec input for completion event")
					}

					isError = streamToolCallComplete(w, flusher, "console_exec", input, response)
				case "str_replace_based_edit_tool":
					var input rishiTools.TextEditorInput
					if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input); err != nil {
						errMsg := fmt.Sprintf("Failed to parse text editor input: %s, error: %v", variant.JSON.Input.Raw(), err)
						log.Error().Err(err).Msgf(errMsg)
					}

					var commandName string
					switch response.(type) {
					case rishiTools.TextEditorViewOutput:
						commandName = string(rishiTools.ViewCommand)
					case rishiTools.TextEditorStrReplaceOutput:
						commandName = string(rishiTools.StrReplaceCommand)
					case rishiTools.TextEditorCreateOutput:
						commandName = string(rishiTools.CreateCommand)
					case rishiTools.TextEditorInsertOutput:
						commandName = string(rishiTools.InsertCommand)
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

// parseAnthropicError converts Anthropic API errors into user-friendly messages
func parseAnthropicError(err error) string {
	errorMsg := err.Error()
	if strings.Contains(errorMsg, "overloaded_error") || strings.Contains(errorMsg, "Overloaded") {
		return "Claude is currently experiencing high demand. Please try again in a few moments."
	}
	return fmt.Sprintf("Claude encountered an error: %v", err)
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
