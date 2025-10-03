package api

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

type consoleExecInput struct {
	Code string `json:"code"`
}

type consoleExecOutput struct {
	Content string `json:"content"`
	Error   string `json:"error"`
}

func consoleExec(input consoleExecInput) consoleExecOutput {
	var output consoleExecOutput

	err := makeToolRequest("/console/exec", input, &output)
	if err != nil {
		log.Error().Err(err).Msg("Failed to call console exec endpoint")
		return consoleExecOutput{
			Error: fmt.Sprintf("Failed to communicate with R server: %v", err),
		}
	}

	return output
}
