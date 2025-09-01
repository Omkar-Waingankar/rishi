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
)

type textEditorController struct {
	safeRoot  string
	wsManager *WebSocketManager
}

type textEditorInput struct {
	Command TextEditorCommand `json:"command"`
	textEditorViewInput
}

type textEditorViewInput struct {
	Path      string `json:"path"`
	ViewRange []int  `json:"view_range"`
}

type textEditorViewOutput struct {
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
