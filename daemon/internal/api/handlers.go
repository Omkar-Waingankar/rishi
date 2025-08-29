package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/rs/zerolog/log"
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

	// System prompt for RStudio-embedded coding agent (chat now; agentic tools soon)
	systemPrompt := `You are **Tibbl**, an embedded coding copilot inside RStudio (Cursor-for-R).

	MISSION
	- Help users write, debug, refactor, and understand code (R-first). Also support Python-in-RStudio, Shiny, Quarto/Rmd, SQL, data science, and package dev.
	- Lead with the solution. Use minimal, runnable code. Be precise, pragmatic, and friendly.

	VOICE & SHAPE
	- Write naturally. Narrate progress briefly ("I'll scan X... now applying Y... done."). Use short bullets or a small heading only when it clarifies; avoid templated sections.
	- Ask a clarifying question only if strictly required. Otherwise state reasonable assumptions and proceed.

	QUESTION HANDLING PROTOCOL
	- **Explain** (concepts, APIs, repos, design): give a crisp overview first, then key details and how to run/verify. Cite important files/symbols when relevant.
	- **Implement** (make/change code): apply minimal edits or self-contained snippets; keep unrelated changes out.
	- **Diagnose** (errors/bugs/perf): restate the symptom, propose likely causes, show one primary fix path; mention one fallback briefly.
	- When multiple viable paths exist, briefly weigh trade-offs and **pick one**.

	TOOL USE (e.g., read_file, list_files)
	- Follow each tool's input schema exactly, paying attention to required fields. Omit optional fields rather than sending nulls; respect ranges.
	- Pay attention to tool errors. If the *same* tool fails twice with the *same* error, stop retrying and pivot to another approach/tool, ask one targeted question, or proceed with a stated assumption.
	- Only use tools to read source files (.R, .Rmd). For data files, write code to inspect/preview shape/columns instead of opening them via tools.

	EDITING & FILE CHANGES
	- Reference files/symbols precisely (e.g., "R/model.R: fit_model()"; include line numbers if known).
	- Before changes, briefly summarize the relevant code/intent you inferred.
	- Provide minimal diffs or well-scoped replacement blocks; include required "library()"/"imports". Keep unrelated changes out.
	- After edits or tool-based modifications, **end with a concise confirmation of what changed and where** (a short "Changes made" list). Keep it factual.

	CODE STYLE & R DEFAULTS
	- Prefer tidyverse for data work; include necessary "library()" calls. For modeling, use tidymodels with a clear split/fit/evaluate flow and "set.seed(123)".
	- Use base R when simpler or zero-deps. Name things clearly; avoid magic numbers; handle edge cases.
	- For large data, suggest sampling or arrow/data.table patterns.
	- Assume an RStudio Project; recommend "renv" when adding packages. Suggest RStudio Jobs for long tasks; use Terminal/Build tools/Connections/Snippets/Addins when relevant.
	- For Quarto/Rmd: chunk options, caching, parameters when helpful. For Shiny: modules, reactive hygiene, testable server logic.

	OUTPUT RULES
	- Keep responses tight. Don't wrap the entire reply in one code block; include only runnable snippets.
	- Truncate huge outputs; show head/tail and give exact commands to reproduce locally.
	- Avoid risky or project-wide changes without explicit permission. Never expose secrets; prefer env vars/config.

	END CONDITIONS (non-negotiable)
	- If tools were used or files were edited: end with a short **Changes made** confirmation (what/where).
	- If there were meaningful options: end with a **Recommendation** (pick the best and say why).
	- Add **Next checks** only when verification helps (exact commands or views to confirm success).
`

	// Convert history into []anthropic.MessageParam, include system prompt, then append latest user message
	var msgs []anthropic.MessageParam
	// Prepend system prompt as a message to keep behavior similar
	msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(systemPrompt)))
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
	if in.Model != "" {
		model = anthropic.Model(in.Model)
	}

	maxTokens := in.MaxTok
	if maxTokens == 0 {
		maxTokens = 1024
	}

	toolParams := []anthropic.ToolParam{
		ReadFileTool,
		ListFilesTool,
	}
	tools := make([]anthropic.ToolUnionParam, len(toolParams))
	for i, toolParam := range toolParams {
		tools[i] = anthropic.ToolUnionParam{OfTool: &toolParam}
	}

	for {
		stream := s.anthropicClient.Messages.NewStreaming(r.Context(), anthropic.MessageNewParams{
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
				_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
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
			_ = json.NewEncoder(w).Encode(map[string]any{"error": fmt.Sprintf("streaming error: %v", err)})
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
				// Stream tool call start event to frontend
				_ = json.NewEncoder(w).Encode(map[string]any{
					"tool_call": map[string]any{
						"name":   block.Name,
						"input":  variant.JSON.Input.Raw(),
						"status": "requesting",
					},
				})
				flusher.Flush()

				var response interface{}
				switch block.Name {
				case "read_file":
					var input ReadFileToolInput
					if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input); err != nil {
						errMsg := fmt.Sprintf("Failed to parse read_file input: %s, error: %v", variant.JSON.Input.Raw(), err)
						log.Error().Err(err).Msgf(errMsg)
						response = ReadFileToolResult{
							Error: errMsg,
						}
						break
					}

					if err := validateReadFileToolInput(input); err != nil {
						errMsg := fmt.Sprintf("Failed to validate read_file input: %s, error: %v", input.Path, err)
						log.Error().Err(err).Msgf(errMsg)
						response = ReadFileToolResult{
							Error: errMsg,
						}
						break
					}

					// Make HTTP call to RStudio frontend to read file
					readResult, err := s.readFileFromRStudio(input.Path)
					if err != nil {
						errMsg := fmt.Sprintf("Failed to read file from RStudio frontend: %s, error: %v", input.Path, err)
						log.Error().Err(err).Msgf(errMsg)
						response = ReadFileToolResult{
							Error: errMsg,
						}
					} else {
						response = *readResult
					}
				case "list_files":
					var input ListFilesToolInput
					if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input); err != nil {
						errMsg := fmt.Sprintf("Failed to parse list_files input: %s, error: %v", variant.JSON.Input.Raw(), err)
						log.Error().Err(err).Msgf(errMsg)
						response = ListFilesToolResult{
							Error: errMsg,
						}
						break
					}

					if err := validateListFilesToolInput(input); err != nil {
						errMsg := fmt.Sprintf("Failed to validate list_files input: path=%s, recursive=%v, error: %v", input.Path, input.Recursive, err)
						log.Error().Err(err).Msgf(errMsg)
						response = ListFilesToolResult{
							Error: errMsg,
						}
						break
					}

					// Make HTTP call to RStudio frontend to list files
					listResult, err := s.listFilesFromRStudio(input.Path, input.Recursive)
					if err != nil {
						errMsg := fmt.Sprintf("Failed to list files from RStudio frontend: path=%s, recursive=%v, error: %v", input.Path, input.Recursive, err)
						log.Error().Err(err).Msgf(errMsg)
						response = ListFilesToolResult{
							Error: errMsg,
						}
					} else {
						response = *listResult
					}
				}

				b, err := json.Marshal(response)
				if err != nil {
					http.Error(w, "error parsing tool result", http.StatusInternalServerError)
					return
				}

				// Stream tool call completion event to frontend
				log.Info().Msgf("tool call completed: %s, result: %s", block.Name, string(b)[:min(100, len(string(b)))])
				_ = json.NewEncoder(w).Encode(map[string]any{
					"tool_call": map[string]any{
						"name":   block.Name,
						"input":  variant.JSON.Input.Raw(),
						"status": "completed",
						"result": string(b),
					},
				})
				flusher.Flush()

				msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewToolUseBlock(block.ID, variant.JSON.Input, block.Name)))

				toolResults = append(toolResults, anthropic.NewToolResultBlock(block.ID, string(b), false))
				msgs = append(msgs, anthropic.NewUserMessage(toolResults...))
			}
		}

		log.Info().Msgf("toolResults: %d", len(toolResults))

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
