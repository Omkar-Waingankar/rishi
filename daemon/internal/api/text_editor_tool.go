package api

import (
	"fmt"

	"github.com/rs/zerolog/log"
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

type textEditorController struct {
	safeRoot  string
	wsManager *WebSocketManager
}

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
	ViewRange []int  `json:"view_range"`
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

func (c *textEditorController) view(input textEditorViewInput) textEditorViewOutput {
	// Forward the request to the R server via WebSocket with proper struct
	response, err := c.wsManager.sendTextEditorCommand(input)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send text editor command to R server")
		return textEditorViewOutput{
			Error: fmt.Sprintf("Failed to communicate with R server: %v", err),
		}
	}

	// Response is already the correct type!
	if response == nil {
		return textEditorViewOutput{
			Error: "No response from R server",
		}
	}

	return *response
}

func (c *textEditorController) strReplace(input textEditorStrReplaceInput) textEditorStrReplaceOutput {
	// Forward the request to the R server via WebSocket with proper struct
	response, err := c.wsManager.sendTextEditorStrReplaceCommand(input)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send text editor str_replace command to R server")
		return textEditorStrReplaceOutput{
			Error: fmt.Sprintf("Failed to communicate with R server: %v", err),
		}
	}

	// Response is already the correct type!
	if response == nil {
		return textEditorStrReplaceOutput{
			Error: "No response from R server",
		}
	}

	return *response
}

func (c *textEditorController) create(input textEditorCreateInput) textEditorCreateOutput {
	// Forward the request to the R server via WebSocket with proper struct
	response, err := c.wsManager.sendTextEditorCreateCommand(input)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send text editor create command to R server")
		return textEditorCreateOutput{
			Error: fmt.Sprintf("Failed to communicate with R server: %v", err),
		}
	}

	// Response is already the correct type!
	if response == nil {
		return textEditorCreateOutput{
			Error: "No response from R server",
		}
	}

	return *response
}

func (c *textEditorController) insert(input textEditorInsertInput) textEditorInsertOutput {
	// Forward the request to the R server via WebSocket with proper struct
	response, err := c.wsManager.sendTextEditorInsertCommand(input)
	if err != nil {
		log.Error().Err(err).Msg("Failed to send text editor insert command to R server")
		return textEditorInsertOutput{
			Error: fmt.Sprintf("Failed to communicate with R server: %v", err),
		}
	}

	// Response is already the correct type!
	if response == nil {
		return textEditorInsertOutput{
			Error: "No response from R server",
		}
	}

	return *response
}
