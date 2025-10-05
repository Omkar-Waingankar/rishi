package api

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"
)

// handleOpenAIChat proxies a streaming request with history to OpenAI and emits NDJSON lines
// of the form {"text": "..."} and a final {"is_final": true}.
func (s *ServerClient) handleOpenAIChat(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get API key from header or config
	apiKey := r.Header.Get("X-Provider-API-Key")
	if apiKey == "" {
		// Try to get from config
		var err error
		apiKey, err = GetOpenAIAPIKey()
		if err != nil || apiKey == "" {
			http.Error(w, "missing X-Provider-API-Key header and no OpenAI API key in config", http.StatusUnauthorized)
			return
		}
	}

	model := openai.GPT4o

	selectedModel := r.Header.Get("X-Model")
	if selectedModel != "" {
		switch selectedModel {
		case "gpt-4o":
			model = openai.GPT4o
		case "gpt-4o-mini":
			model = openai.GPT4oMini
		default:
			model = openai.GPT4o
		}
	}

	// Create OpenAI client
	client := openai.NewClient(apiKey)

	type inboundMessage struct {
		Role    string           `json:"role"`
		Content []inboundContent `json:"content"`
	}
	type reqBody struct {
		History []inboundMessage `json:"history"`
		Content []inboundContent `json:"content"`
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

	// Convert history to OpenAI format
	var messages []openai.ChatCompletionMessage

	// Add system message
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: RISHI_SYSTEM_PROMPT,
	})

	// Add conversation history
	for _, m := range in.History {
		var messageContent string

		switch m.Role {
		case "user":
			messageContent = convertToOpenaiContent(m.Content)
			if messageContent != "" {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: messageContent,
				})
			}
		case "assistant":
			messageContent = convertToOpenaiContent(m.Content)
			if messageContent != "" {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: messageContent,
				})
			}
		}
	}

	// Add current user message
	currentMessageContent := convertToOpenaiContent(in.Content)
	if currentMessageContent != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: currentMessageContent,
		})
	}

	maxTokens := in.MaxTok
	if maxTokens == 0 {
		maxTokens = defaultMaxTokens
	}

	// Create streaming request
	req := openai.ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: 0.1,
		Stream:      true,
	}

	stream, err := client.CreateChatCompletionStream(r.Context(), req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create OpenAI stream")
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		flusher.Flush()
		return
	}
	defer stream.Close()

	// Stream responses
	for {
		response, err := stream.Recv()
		if err != nil {
			if err.Error() == "stream finished" {
				break
			}
			log.Error().Err(err).Msg("OpenAI stream error")
			_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
			flusher.Flush()
			return
		}

		// Send text content
		if len(response.Choices) > 0 && response.Choices[0].Delta.Content != "" {
			_ = json.NewEncoder(w).Encode(map[string]any{"text": response.Choices[0].Delta.Content})
			flusher.Flush()
		}
	}

	// Send final message
	_ = json.NewEncoder(w).Encode(map[string]any{"is_final": true})
	flusher.Flush()
}
