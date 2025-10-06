package api

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

type rHelpInput struct {
	Package string  `json:"package"`
	Topic   *string `json:"topic,omitempty"`
}

type rHelpOutput struct {
	Content string `json:"content"`
	Error   string `json:"error"`
}

func rHelp(input rHelpInput) rHelpOutput {
	var output rHelpOutput

	err := makeToolRequest("/r_help", input, &output)
	if err != nil {
		log.Error().Err(err).Msg("Failed to call r_help endpoint")
		return rHelpOutput{
			Error: fmt.Sprintf("Failed to communicate with R server: %v", err),
		}
	}

	return output
}
