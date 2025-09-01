package api

import "errors"

var (
	// ErrClientNotConnected is returned when trying to send a message to a client that is not connected
	ErrClientNotConnected = errors.New("client not connected")
	
	// ErrTimeout is returned when a request times out
	ErrTimeout = errors.New("request timeout")
)