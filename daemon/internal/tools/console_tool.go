package tools

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

type ConsoleExecInput struct {
	Code string `json:"code"`
}

type ConsoleExecOutput struct {
	Content string `json:"content"`
	Error   string `json:"error"`
}

func ConsoleExec(input ConsoleExecInput) ConsoleExecOutput {
	var output ConsoleExecOutput

	err := makeToolRequest("/console/exec", input, &output)
	if err != nil {
		log.Error().Err(err).Msg("Failed to call console exec endpoint")
		return ConsoleExecOutput{
			Error: fmt.Sprintf("Failed to communicate with R server: %v", err),
		}
	}

	return output
}
