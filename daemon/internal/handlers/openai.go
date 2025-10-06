package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	rishiTools "github.com/Omkar-Waingankar/rishi/daemon/internal/tools"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// handleOpenAIChat proxies a streaming request with history to OpenAI and emits NDJSON lines
// of the form {"text": "..."} and a final {"is_final": true}.
func HandleOpenAIChat(w http.ResponseWriter, r *http.Request) {
	if !validateChatRequest(w, r) {
		return
	}

	// Get API key from header
	apiKey, ok := validateAPIKey(w, r)
	if !ok {
		return
	}

	// Create OpenAI client
	client := openai.NewClient(apiKey)

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

	// Convert history to OpenAI format
	var messages []openai.ChatCompletionMessage

	// Add system message
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: RISHI_SYSTEM_PROMPT,
	})

	// Add conversation history
	for i, m := range in.History {
		switch m.Role {
		case "user":
			// Convert content blocks for user messages (supports text + images)
			contentParts, err := convertToOpenAIContentParts(m.Content)
			if err != nil {
				log.Error().Err(err).Msgf("Error converting user history content")
				http.Error(w, fmt.Sprintf("Invalid user message content: %v", err), http.StatusBadRequest)
				return
			}
			if len(contentParts) > 0 {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:         openai.ChatMessageRoleUser,
					MultiContent: contentParts,
				})
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
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: textContent,
				})
			}
		default:
			// ignore
		}
		log.Info().Msgf("history message %d: role %s, %d content blocks, content: %s", i, m.Role, len(m.Content), formatMessageContentArrayForDebug(m.Content))
	}

	// Handle the new user message content
	if len(in.Content) > 0 {
		contentParts, err := convertToOpenAIContentParts(in.Content)
		if err != nil {
			log.Error().Err(err).Msgf("Error converting user message content")
			http.Error(w, fmt.Sprintf("Invalid message content: %v", err), http.StatusBadRequest)
			return
		}
		if len(contentParts) > 0 {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:         openai.ChatMessageRoleUser,
				MultiContent: contentParts,
			})
		}
	}
	log.Info().Msgf("new user message: %d content blocks, content: %s", len(in.Content), formatMessageContentArrayForDebug(in.Content))

	// Get model from X-Model header
	model := openai.GPT4o
	selectedModel := r.Header.Get("X-Model")
	if selectedModel != "" {
		switch selectedModel {
		case "gpt-4o":
			model = openai.GPT4o
			log.Info().Msgf("Using GPT-4o model")
		case "gpt-4o-mini":
			model = openai.GPT4oMini
			log.Info().Msgf("Using GPT-4o-mini model")
		case "gpt-5":
			model = openai.GPT5
			log.Info().Msgf("Using GPT-5 model")
		default:
			log.Warn().Msgf("Unknown model requested: %s, using default GPT-4o", selectedModel)
		}
	} else {
		log.Info().Msgf("No model specified, using default GPT-4o")
	}

	maxTokens := in.MaxTok
	if maxTokens == 0 {
		maxTokens = defaultMaxTokens
	}

	// Define tools for OpenAI
	tools := []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "console_exec",
				Description: "Executes R code in the user's R console. The code will be sent to the console and executed immediately.\n\nParameters:\n- code (required): The R code to execute in the console.\n\nExample:\n- Execute code: {\"code\": \"print('Hello, World!')\"}",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"code": {
							Type:        jsonschema.String,
							Description: "The R code to execute in the console",
						},
					},
					Required: []string{"code"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "r_help",
				Description: "Retrieves R package documentation. Use this to look up package-level documentation or specific function/topic documentation when writing R code or answering questions about R packages and their functionality.\n\nParameters:\n- package (required): The name of the R package to get documentation for (e.g., 'dplyr', 'ggplot2', 'rstudioapi').\n- topic (optional): The specific function, method, or topic within the package to get detailed documentation for (e.g., 'addTheme', 'mutate').\n\nExamples:\n- Package-level help: {\"package\": \"ggplot2\"}\n- Topic-specific help: {\"package\": \"rstudioapi\", \"topic\": \"addTheme\"} - equivalent to help(\"addTheme\", package=\"rstudioapi\")",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"package": {
							Type:        jsonschema.String,
							Description: "The name of the R package to get documentation for",
						},
						"topic": {
							Type:        jsonschema.String,
							Description: "The specific function, method, or topic within the package",
						},
					},
					Required: []string{"package"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "text_editor_view",
				Description: "Views the contents of a file or directory. Use this to read files or list directory contents when writing R code or exploring the project structure.\n\nParameters:\n- path (required): The path to the file or directory to view.\n- view_range (optional): Array of two integers [start_line, end_line] to view specific line range.\n\nExamples:\n- View entire file: {\"path\": \"analysis.R\"}\n- View specific lines: {\"path\": \"analysis.R\", \"view_range\": [1, 50]}\n- List directory: {\"path\": \".\"}",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"path": {
							Type:        jsonschema.String,
							Description: "The path to the file or directory to view",
						},
						"view_range": {
							Type:        jsonschema.Array,
							Description: "Optional array of two integers [start_line, end_line] to view specific line range",
							Items: &jsonschema.Definition{
								Type: jsonschema.Integer,
							},
						},
					},
					Required: []string{"path"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "text_editor_str_replace",
				Description: "Replaces exact text in a file. The old_str must match exactly (including whitespace and indentation) for the replacement to succeed.\n\nParameters:\n- path (required): The path to the file to edit.\n- old_str (required): The exact text to find and replace (must match exactly).\n- new_str (required): The text to replace it with.\n\nExample:\n- Replace text: {\"path\": \"script.R\", \"old_str\": \"x <- 1\", \"new_str\": \"x <- 2\"}",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"path": {
							Type:        jsonschema.String,
							Description: "The path to the file to edit",
						},
						"old_str": {
							Type:        jsonschema.String,
							Description: "The exact text to find and replace",
						},
						"new_str": {
							Type:        jsonschema.String,
							Description: "The text to replace it with",
						},
					},
					Required: []string{"path", "old_str", "new_str"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "text_editor_create",
				Description: "Creates a new file with the specified content. Will fail if the file already exists.\n\nParameters:\n- path (required): The path where the new file should be created.\n- file_text (required): The complete content for the new file.\n\nExample:\n- Create file: {\"path\": \"new_script.R\", \"file_text\": \"# My new R script\\nx <- 1\\nprint(x)\"}",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"path": {
							Type:        jsonschema.String,
							Description: "The path where the new file should be created",
						},
						"file_text": {
							Type:        jsonschema.String,
							Description: "The complete content for the new file",
						},
					},
					Required: []string{"path", "file_text"},
				},
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "text_editor_insert",
				Description: "Inserts text at a specific line number in a file. The new text will be inserted before the specified line.\n\nParameters:\n- path (required): The path to the file to edit.\n- insert_line (required): The line number where text should be inserted (1-indexed).\n- new_str (required): The text to insert.\n\nExample:\n- Insert at line 5: {\"path\": \"script.R\", \"insert_line\": 5, \"new_str\": \"# New comment\\n\"}",
				Parameters: jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"path": {
							Type:        jsonschema.String,
							Description: "The path to the file to edit",
						},
						"insert_line": {
							Type:        jsonschema.Integer,
							Description: "The line number where text should be inserted (1-indexed)",
						},
						"new_str": {
							Type:        jsonschema.String,
							Description: "The text to insert",
						},
					},
					Required: []string{"path", "insert_line", "new_str"},
				},
			},
		},
	}

	// Agentic loop for tool calling
	for {
		// Create streaming request
		req := openai.ChatCompletionRequest{
			Model:               model,
			Messages:            messages,
			MaxCompletionTokens: maxTokens,
			Temperature:         0.1,
			Stream:              true,
			Tools:               tools,
		}

		stream, err := client.CreateChatCompletionStream(r.Context(), req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create OpenAI stream")
			_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
			flusher.Flush()
			return
		}

		// Accumulate the full assistant message
		var fullContent string
		var toolCalls []openai.ToolCall

		// Stream responses
		for {
			response, err := stream.Recv()
			if err != nil {
				stream.Close()
				if err.Error() == "stream finished" || err.Error() == "EOF" {
					break
				}
				log.Error().Err(err).Msg("OpenAI stream error")
				_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
				flusher.Flush()
				return
			}

			if len(response.Choices) > 0 {
				delta := response.Choices[0].Delta

				// Send text content
				if delta.Content != "" {
					fullContent += delta.Content
					_ = json.NewEncoder(w).Encode(map[string]any{"text": delta.Content})
					flusher.Flush()
				}

				// Accumulate tool calls
				if len(delta.ToolCalls) > 0 {
					for _, tc := range delta.ToolCalls {
						// Find or create the tool call in our accumulator
						if tc.Index != nil {
							idx := *tc.Index
							// Ensure we have enough slots
							for len(toolCalls) <= idx {
								toolCalls = append(toolCalls, openai.ToolCall{})
							}
							// Accumulate the tool call data
							if tc.ID != "" {
								toolCalls[idx].ID = tc.ID
							}
							if tc.Type != "" {
								toolCalls[idx].Type = tc.Type
							}
							if tc.Function.Name != "" {
								toolCalls[idx].Function.Name = tc.Function.Name
							}
							if tc.Function.Arguments != "" {
								toolCalls[idx].Function.Arguments += tc.Function.Arguments
							}
						}
					}
				}
			}
		}
		stream.Close()

		// Add assistant's response to messages
		if fullContent != "" {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: fullContent,
			})
		}

		// If no tool calls, we're done
		if len(toolCalls) == 0 {
			_ = json.NewEncoder(w).Encode(map[string]any{"is_final": true})
			flusher.Flush()
			break
		}

		// Process tool calls
		var toolMessages []openai.ChatCompletionMessage
		assistantMessage := openai.ChatCompletionMessage{
			Role:      openai.ChatMessageRoleAssistant,
			ToolCalls: toolCalls,
		}
		messages = append(messages, assistantMessage)

		for _, toolCall := range toolCalls {
			log.Info().Msgf("tool use: %s, input: %s", toolCall.Function.Name, toolCall.Function.Arguments)

			var response interface{}
			var isError bool

			switch toolCall.Function.Name {
			case "r_help":
				var input rishiTools.RHelpInput
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
					errMsg := fmt.Sprintf("Failed to parse r_help input: %s, error: %v", toolCall.Function.Arguments, err)
					log.Error().Err(err).Msgf(errMsg)
					response = rishiTools.RHelpOutput{
						Error: errMsg,
					}
				} else {
					streamToolCallStart(w, flusher, "r_help", input)
					response = rishiTools.RHelp(input)
				}

				b, err := json.Marshal(response)
				if err != nil {
					http.Error(w, "error parsing tool result", http.StatusInternalServerError)
					return
				}
				log.Info().Msgf("tool call completed: r_help, result length: %d, result: %s", len(string(b)), string(b)[:min(100, len(string(b)))])
				isError = streamToolCallComplete(w, flusher, "r_help", input, response)

				toolMessages = append(toolMessages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    string(b),
					ToolCallID: toolCall.ID,
				})

			case "console_exec":
				var input rishiTools.ConsoleExecInput
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
					errMsg := fmt.Sprintf("Failed to parse console exec input: %s, error: %v", toolCall.Function.Arguments, err)
					log.Error().Err(err).Msgf(errMsg)
					response = rishiTools.ConsoleExecOutput{
						Error: errMsg,
					}
				} else {
					streamToolCallStart(w, flusher, "console_exec", input)
					response = rishiTools.ConsoleExec(input)
				}

				b, err := json.Marshal(response)
				if err != nil {
					http.Error(w, "error parsing tool result", http.StatusInternalServerError)
					return
				}
				log.Info().Msgf("tool call completed: console_exec, result length: %d, result: %s", len(string(b)), string(b)[:min(100, len(string(b)))])
				isError = streamToolCallComplete(w, flusher, "console_exec", input, response)

				toolMessages = append(toolMessages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    string(b),
					ToolCallID: toolCall.ID,
				})

			case "text_editor_view":
				var input rishiTools.TextEditorViewInput
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
					errMsg := fmt.Sprintf("Failed to parse text editor view input: %s, error: %v", toolCall.Function.Arguments, err)
					log.Error().Err(err).Msgf(errMsg)
					response = rishiTools.TextEditorViewOutput{
						Error: errMsg,
					}
				} else {
					streamToolCallStart(w, flusher, "view", input)
					response = rishiTools.TextEditorView(input)
				}

				b, err := json.Marshal(response)
				if err != nil {
					http.Error(w, "error parsing tool result", http.StatusInternalServerError)
					return
				}
				log.Info().Msgf("tool call completed: text_editor_view, result length: %d, result: %s", len(string(b)), string(b)[:min(100, len(string(b)))])
				isError = streamToolCallComplete(w, flusher, "view", input, response)

				toolMessages = append(toolMessages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    string(b),
					ToolCallID: toolCall.ID,
				})

			case "text_editor_str_replace":
				var input rishiTools.TextEditorStrReplaceInput
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
					errMsg := fmt.Sprintf("Failed to parse text editor str_replace input: %s, error: %v", toolCall.Function.Arguments, err)
					log.Error().Err(err).Msgf(errMsg)
					response = rishiTools.TextEditorStrReplaceOutput{
						Error: errMsg,
					}
				} else {
					streamToolCallStart(w, flusher, "str_replace", input)
					response = rishiTools.TextEditorStrReplace(input)
				}

				b, err := json.Marshal(response)
				if err != nil {
					http.Error(w, "error parsing tool result", http.StatusInternalServerError)
					return
				}
				log.Info().Msgf("tool call completed: text_editor_str_replace, result length: %d, result: %s", len(string(b)), string(b)[:min(100, len(string(b)))])
				isError = streamToolCallComplete(w, flusher, "str_replace", input, response)

				toolMessages = append(toolMessages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    string(b),
					ToolCallID: toolCall.ID,
				})

			case "text_editor_create":
				var input rishiTools.TextEditorCreateInput
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
					errMsg := fmt.Sprintf("Failed to parse text editor create input: %s, error: %v", toolCall.Function.Arguments, err)
					log.Error().Err(err).Msgf(errMsg)
					response = rishiTools.TextEditorCreateOutput{
						Error: errMsg,
					}
				} else {
					streamToolCallStart(w, flusher, "create", input)
					response = rishiTools.TextEditorCreate(input)
				}

				b, err := json.Marshal(response)
				if err != nil {
					http.Error(w, "error parsing tool result", http.StatusInternalServerError)
					return
				}
				log.Info().Msgf("tool call completed: text_editor_create, result length: %d, result: %s", len(string(b)), string(b)[:min(100, len(string(b)))])
				isError = streamToolCallComplete(w, flusher, "create", input, response)

				toolMessages = append(toolMessages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    string(b),
					ToolCallID: toolCall.ID,
				})

			case "text_editor_insert":
				var input rishiTools.TextEditorInsertInput
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
					errMsg := fmt.Sprintf("Failed to parse text editor insert input: %s, error: %v", toolCall.Function.Arguments, err)
					log.Error().Err(err).Msgf(errMsg)
					response = rishiTools.TextEditorInsertOutput{
						Error: errMsg,
					}
				} else {
					streamToolCallStart(w, flusher, "insert", input)
					response = rishiTools.TextEditorInsert(input)
				}

				b, err := json.Marshal(response)
				if err != nil {
					http.Error(w, "error parsing tool result", http.StatusInternalServerError)
					return
				}
				log.Info().Msgf("tool call completed: text_editor_insert, result length: %d, result: %s", len(string(b)), string(b)[:min(100, len(string(b)))])
				isError = streamToolCallComplete(w, flusher, "insert", input, response)

				toolMessages = append(toolMessages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    string(b),
					ToolCallID: toolCall.ID,
				})
			}

			// If there was an error, mark it in the tool result
			if isError {
				log.Warn().Msgf("Tool call %s returned an error", toolCall.Function.Name)
			}
		}

		// Add all tool results to messages
		messages = append(messages, toolMessages...)
	}
}

// convertToOpenAIContentParts converts inbound content to OpenAI message parts
func convertToOpenAIContentParts(contents []inboundContent) ([]openai.ChatMessagePart, error) {
	var parts []openai.ChatMessagePart

	for _, content := range contents {
		switch content.Type {
		case "text":
			if content.Content != "" {
				parts = append(parts, openai.ChatMessagePart{
					Type: openai.ChatMessagePartTypeText,
					Text: content.Content,
				})
			}
		case "image":
			if err := validateImageContent(content); err != nil {
				return nil, fmt.Errorf("invalid image content: %v", err)
			}

			// OpenAI expects data URI format for base64 images
			dataURI := fmt.Sprintf("data:%s;base64,%s", content.MediaType, content.DataBase64)
			parts = append(parts, openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL: dataURI,
				},
			})
		default:
			log.Warn().Msgf("Unknown content type: %s", content.Type)
		}
	}

	return parts, nil
}
