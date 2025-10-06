package tools

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

type RHelpInput struct {
	Package string  `json:"package"`
	Topic   *string `json:"topic,omitempty"`
}

type RHelpOutput struct {
	Content string `json:"content"`
	Error   string `json:"error"`
}

func RHelp(input RHelpInput) RHelpOutput {
	var output RHelpOutput

	err := makeToolRequest("/r_help", input, &output)
	if err != nil {
		log.Error().Err(err).Msg("Failed to call r_help endpoint")
		return RHelpOutput{
			Error: fmt.Sprintf("Failed to communicate with R server: %v", err),
		}
	}

	return output
}
