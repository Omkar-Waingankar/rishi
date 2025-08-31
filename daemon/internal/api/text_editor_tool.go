package api

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// TextEditorCommand represents the available commands for the text editor tool
type TextEditorCommand string

const (
	// ViewCommand represents the view command for reading file contents
	ViewCommand TextEditorCommand = "view"
)

type textEditorController struct {
	safeRoot string
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
	relativePath := input.Path
	viewRange := input.ViewRange

	// TODO: check if file exists, etc.
	if input.Path == "" {
		return textEditorViewOutput{
			Error: "Path is required",
		}
	}

	absolutePath := filepath.Join(c.safeRoot, relativePath)
	var result bytes.Buffer

	info, err := os.Stat(absolutePath)
	if err != nil {
		log.Error().Err(err).Msgf("failed to stat file '%s'", relativePath)
		return textEditorViewOutput{
			Error: "Failed to stat file",
		}
	}

	if info.IsDir() {
		// List directory contents
		entries, err := os.ReadDir(absolutePath)
		if err != nil {
			log.Error().Err(err).Msgf("failed to read directory '%s'", relativePath)
			return textEditorViewOutput{
				Error: "Failed to read directory",
			}
		}

		result.WriteString(fmt.Sprintf("Directory listing for '%s':\n", relativePath))

		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() {
				// Mark directories with trailing slash
				result.WriteString(fmt.Sprintf("%s/\n", name))
			} else {
				result.WriteString(fmt.Sprintf("%s\n", name))
			}
		}
	} else {
		// Read the file contents
		content, err := os.ReadFile(absolutePath)
		if err != nil {
			log.Error().Err(err).Msgf("failed to read file '%s'", relativePath)
			return textEditorViewOutput{
				Error: "Failed to read file",
			}
		}

		// Convert content to string and split into lines
		lines := bytes.Split(content, []byte("\n"))

		// Determine which lines to include based on viewRange
		startLine := 1
		endLine := len(lines)

		if len(viewRange) == 2 {
			startLine = viewRange[0]
			endLine = viewRange[1]
		}

		// Build the result with line numbers
		result.WriteString(fmt.Sprintf("File contents for '%s':\n", relativePath))

		for i := startLine - 1; i < endLine; i++ {
			lineNum := i + 1
			lineContent := string(lines[i])
			result.WriteString(fmt.Sprintf("%d: %s\n", lineNum, lineContent))
		}
	}

	return textEditorViewOutput{
		Content: result.String(),
	}
}
