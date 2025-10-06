package tools

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

type TextEditorInput struct {
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

type TextEditorViewInput struct {
	Path      string `json:"path"`
	ViewRange []int  `json:"view_range,omitempty"`
}

type TextEditorViewOutput struct {
	Content string `json:"content"`
	Error   string `json:"error"`
}

type TextEditorStrReplaceInput struct {
	Path   string `json:"path"`
	OldStr string `json:"old_str"`
	NewStr string `json:"new_str"`
}

type TextEditorStrReplaceOutput struct {
	Content string `json:"content"`
	Error   string `json:"error"`
}

type TextEditorCreateInput struct {
	Path     string `json:"path"`
	FileText string `json:"file_text"`
}

type TextEditorCreateOutput struct {
	Content string `json:"content"`
	Error   string `json:"error"`
}

type TextEditorInsertInput struct {
	Path       string `json:"path"`
	InsertLine int    `json:"insert_line"`
	NewStr     string `json:"new_str"`
}

type TextEditorInsertOutput struct {
	Content string `json:"content"`
	Error   string `json:"error"`
}

func TextEditorView(input TextEditorViewInput) TextEditorViewOutput {
	var output TextEditorViewOutput

	err := makeToolRequest("/text_editor/view", input, &output)
	if err != nil {
		log.Error().Err(err).Msg("Failed to call text editor view endpoint")
		return TextEditorViewOutput{
			Error: fmt.Sprintf("Failed to communicate with R server: %v", err),
		}
	}

	return output
}

func TextEditorStrReplace(input TextEditorStrReplaceInput) TextEditorStrReplaceOutput {
	var output TextEditorStrReplaceOutput

	err := makeToolRequest("/text_editor/str_replace", input, &output)
	if err != nil {
		log.Error().Err(err).Msg("Failed to call text editor str_replace endpoint")
		return TextEditorStrReplaceOutput{
			Error: fmt.Sprintf("Failed to communicate with R server: %v", err),
		}
	}

	return output
}

func TextEditorCreate(input TextEditorCreateInput) TextEditorCreateOutput {
	var output TextEditorCreateOutput

	err := makeToolRequest("/text_editor/create", input, &output)
	if err != nil {
		log.Error().Err(err).Msg("Failed to call text editor create endpoint")
		return TextEditorCreateOutput{
			Error: fmt.Sprintf("Failed to communicate with R server: %v", err),
		}
	}

	return output
}

func TextEditorInsert(input TextEditorInsertInput) TextEditorInsertOutput {
	var output TextEditorInsertOutput

	err := makeToolRequest("/text_editor/insert", input, &output)
	if err != nil {
		log.Error().Err(err).Msg("Failed to call text editor insert endpoint")
		return TextEditorInsertOutput{
			Error: fmt.Sprintf("Failed to communicate with R server: %v", err),
		}
	}

	return output
}
