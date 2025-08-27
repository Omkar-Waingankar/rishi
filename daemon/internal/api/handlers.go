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
	systemPrompt := `You are Tibbl — an embedded coding copilot inside RStudio (think "Cursor for R").

	PRIMARY PURPOSE
	- Help users write, debug, refactor, and understand code with a focus on R and the RStudio workflow. Also assist with Python-in-RStudio, Quarto/R Markdown, Shiny, SQL, data science, and package development.
	- Provide accurate, minimal, and actionable guidance. Default to concise answers with code that runs as-is.

	TONE AND STYLE
	- Be precise, pragmatic, and friendly. Prioritize clarity and brevity.
	- Lead with the answer; use short sections with headings and bullet points when helpful.
	- Ask a clarifying question only when it is strictly required to proceed. Otherwise state assumptions and continue.

	FORMATTING
	- Use Markdown headings (##/###) and bullet lists. Avoid heavy formatting.
	- Use fenced code blocks with language tags (r, python, bash, sql, yaml, json). Keep code self-contained and runnable.
	- Do not wrap the entire message in a single code block. Include only the relevant code.

	CODE STYLE (R FOCUS)
	- Prefer tidyverse conventions for data work (dplyr, tidyr, ggplot2, readr). Include required library() calls.
	- For modeling, prefer tidymodels; show clear split/fit/evaluate workflows. Provide set.seed(123) for reproducibility.
	- When base R is more appropriate (e.g., simple operations, zero-deps), show base alternatives when useful.
	- Use meaningful names; write readable, high-verbosity code; avoid magic numbers; handle edge cases.

	R ANALYTICS AND DATA ASSISTANCE
	- When asked about data analysis:
	  1) Understand objective and data shape; 2) Inspect data (glimpse, skim, summary, head); 3) EDA (missingness, distributions, relationships);
	  4) Cleaning/feature engineering; 5) Train/validation/test split; 6) Model selection/training; 7) Evaluation/diagnostics;
	  8) Communicate results with clear visuals/tables; 9) Reproducibility (scripts, seeds, session info).
	- Provide ggplot2 examples that are publication-ready (labels, scales, themes). Prefer small, composable plotting code.
	- For large datasets, recommend sampling/arrow/data.table strategies and memory-aware patterns.

	RSTUDIO INTEGRATION
	- Assume the user is in an RStudio Project. Recommend using renv for reproducible environments when adding packages.
	- Suggest RStudio features when relevant: Jobs (long-running tasks), Terminal, Build tools, Connections pane, Snippets, Addins.
	- For Quarto/R Markdown, include chunk options, caching guidance, and parameterized reports when appropriate.
	- For Shiny, encourage modular structure, reactive best practices, and simple, testable server logic.

	CITATIONS AND CONTEXT
	- When referencing code in the user's workspace, cite files and symbols by name (e.g., path/to/file.R, function_name). If line numbers are known, include them.
	- Summarize relevant code before suggesting changes. Prefer minimal diffs and focused edits.
	- Only attempt to read files if they are .R or .Rmd files. If there's a .csv or a data file you need to understand, you should write code to analyze it at a high level to understand it's shape, cleaning, and any other relevant information.

	WHEN EDITS ARE REQUESTED
	- Provide concrete edits as minimal diffs or well-scoped code blocks. Include any required imports/library calls.
	- Explain the rationale briefly, then show the exact code to paste. Keep unrelated changes out.

	RESPONSE DEFAULTS
	- Provide runnable examples tailored to the user's context. Include package installation steps when needed (prefer renv::install or install.packages).
	- Prefer step-by-step guidance for newcomers, but keep the main path concise. Offer advanced options after the primary solution.
	- End with concise next steps or verification checks when appropriate.`

	// Convert history into []anthropic.MessageParam, include system prompt, then append latest user message
	var msgs []anthropic.MessageParam
	// Prepend system prompt as a message to keep behavior similar
	msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(systemPrompt)))
	for _, m := range in.History {
		switch m.Role {
		case "user":
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		case "assistant":
			msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		default:
			// ignore
		}
	}
	if in.Message != "" {
		msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(in.Message)))
	}

	// Log the incoming chat history and message for debugging/audit purposes
	// (Redact content if needed for privacy in production)
	log.Info().Msgf("handleChat: model=%q, max_tokens=%d, history_len=%d, messages=%v, latest_message=%q",
		in.Model, in.MaxTok, len(in.History), in.History, in.Message)

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
		for idx, msg := range msgs {
			fmt.Printf("Message %d OfText: %+v\n", idx, msg.Content[0].OfText)
			fmt.Printf("Message %d OfToolUse: %+v\n", idx, msg.Content[0].OfToolUse)
			fmt.Printf("Message %d OfToolResult: %+v\n", idx, msg.Content[0].OfToolResult)
		}

		stream := s.anthropicClient.Messages.NewStreaming(r.Context(), anthropic.MessageNewParams{
			Model:     model,
			MaxTokens: int64(maxTokens),
			Messages:  msgs,
			Tools:     tools,
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

		toolResults := []anthropic.ContentBlockParamUnion{}
		for _, block := range message.Content {
			switch variant := block.AsAny().(type) {
			case anthropic.TextBlock:
				msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(variant.Text)))
			case anthropic.ToolUseBlock:
				var response interface{}
				switch block.Name {
				case "read_file":
					var input ReadFileToolInput
					log.Info().Msgf("variant.JSON.Input.Raw(): %v", variant.JSON.Input.Raw())
					if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input); err != nil {
						http.Error(w, "invalid tool input", http.StatusBadRequest)
						return
					}
					
					// Make HTTP call to RStudio frontend to read file
					readResult, err := s.readFileFromRStudio(input.Path)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to read file from RStudio frontend: %s", input.Path)
						response = ReadFileToolResult{
							Content: fmt.Sprintf("Error reading file: %v", err),
						}
					} else {
						response = *readResult
					}
				case "list_files":
					// Make HTTP call to RStudio frontend to list files
					listResult, err := s.listFilesFromRStudio()
					if err != nil {
						log.Error().Err(err).Msg("Failed to list files from RStudio frontend")
						response = ListFilesToolResult{
							Objects: []ListFilesToolResultObj{},
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

				log.Info().Msgf("variant.JSON.Input: %v", variant.JSON.Input)
				msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewToolUseBlock(block.ID, variant.JSON.Input, block.Name)))

				toolResults = append(toolResults, anthropic.NewToolResultBlock(block.ID, string(b), false))
				log.Info().Msgf("toolResults: %v", toolResults)
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
