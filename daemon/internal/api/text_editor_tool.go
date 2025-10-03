package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// R tool server configuration
	rToolServerPort    = "8082"
	rToolServerTimeout = 30 * time.Second
)

// TextEditorCommand represents the available commands for the text editor tool
type TextEditorCommand string

const (
	// ViewCommand represents the view command for reading file contents
	ViewCommand TextEditorCommand = "view"
	// StrReplaceCommand represents the str_replace command for replacing text
	StrReplaceCommand TextEditorCommand = "str_replace"
	// CreateCommand represents the create command for creating new files
	CreateCommand TextEditorCommand = "create"
	// InsertCommand represents the insert command for inserting text at a specific line
	InsertCommand TextEditorCommand = "insert"
)

type textEditorInput struct {
	Command TextEditorCommand `json:"command"`

	// Common fields
	Path string `json:"path"`

	// View-specific fields
	ViewRange []int `json:"view_range"`

	// StrReplace-specific fields
	OldStr string `json:"old_str"`

	// StrReplace and Insert shared fields
	NewStr string `json:"new_str"`

	// Create-specific fields
	FileText string `json:"file_text"`

	// Insert-specific fields
	InsertLine int    `json:"insert_line"`
	InsertText string `json:"insert_text"` // Actual field name Anthropic sends (despite docs)
}

type textEditorViewInput struct {
	Path      string `json:"path"`
	ViewRange []int  `json:"view_range,omitempty"`
}

type textEditorViewOutput struct {
	Content string `json:"content"`
	Error   string `json:"error"`
}

type textEditorStrReplaceInput struct {
	Path   string `json:"path"`
	OldStr string `json:"old_str"`
	NewStr string `json:"new_str"`
}

type textEditorStrReplaceOutput struct {
	Content string `json:"content"`
	Error   string `json:"error"`
}

type textEditorCreateInput struct {
	Path     string `json:"path"`
	FileText string `json:"file_text"`
}

type textEditorCreateOutput struct {
	Content string `json:"content"`
	Error   string `json:"error"`
}

type textEditorInsertInput struct {
	Path       string `json:"path"`
	InsertLine int    `json:"insert_line"`
	NewStr     string `json:"new_str"`
}

type textEditorInsertOutput struct {
	Content string `json:"content"`
	Error   string `json:"error"`
}

// HTTP client for R tool server
var toolClient = &http.Client{
	Timeout: rToolServerTimeout,
}

// makeToolRequest makes an HTTP POST request to the R tool server
func makeToolRequest(endpoint string, payload interface{}, response interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:%s%s", rToolServerPort, endpoint), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := toolClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	return nil
}

func textEditorView(input textEditorViewInput) textEditorViewOutput {
	var output textEditorViewOutput

	err := makeToolRequest("/text_editor/view", input, &output)
	if err != nil {
		log.Error().Err(err).Msg("Failed to call text editor view endpoint")
		return textEditorViewOutput{
			Error: fmt.Sprintf("Failed to communicate with R server: %v", err),
		}
	}

	return output
}

func textEditorStrReplace(input textEditorStrReplaceInput) textEditorStrReplaceOutput {
	var output textEditorStrReplaceOutput

	err := makeToolRequest("/text_editor/str_replace", input, &output)
	if err != nil {
		log.Error().Err(err).Msg("Failed to call text editor str_replace endpoint")
		return textEditorStrReplaceOutput{
			Error: fmt.Sprintf("Failed to communicate with R server: %v", err),
		}
	}

	return output
}

func textEditorCreate(input textEditorCreateInput) textEditorCreateOutput {
	var output textEditorCreateOutput

	err := makeToolRequest("/text_editor/create", input, &output)
	if err != nil {
		log.Error().Err(err).Msg("Failed to call text editor create endpoint")
		return textEditorCreateOutput{
			Error: fmt.Sprintf("Failed to communicate with R server: %v", err),
		}
	}

	return output
}

func textEditorInsert(input textEditorInsertInput) textEditorInsertOutput {
	var output textEditorInsertOutput

	err := makeToolRequest("/text_editor/insert", input, &output)
	if err != nil {
		log.Error().Err(err).Msg("Failed to call text editor insert endpoint")
		return textEditorInsertOutput{
			Error: fmt.Sprintf("Failed to communicate with R server: %v", err),
		}
	}

	return output
}
