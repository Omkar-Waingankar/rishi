package api

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// sendTextEditorCommand sends a text editor view command to the R client and waits for response
func (m *WebSocketManager) sendTextEditorCommand(input textEditorViewInput) (*textEditorViewOutput, error) {
	// Create response channel
	responseChan := make(chan *textEditorViewOutput, 1)

	// Send the tool command
	requestID, err := m.sendToolCommand("text_editor", "view", input, responseChan)
	if err != nil {
		close(responseChan)
		return nil, err
	}

	// Clean up function
	cleanup := func() {
		m.requestMu.Lock()
		delete(m.pendingRequests, requestID)
		m.requestMu.Unlock()
		close(responseChan)
	}

	// Wait for response with timeout
	select {
	case response := <-responseChan:
		cleanup()
		return response, nil
	case <-time.After(30 * time.Second):
		cleanup()
		return nil, ErrTimeout
	}
}

// sendTextEditorStrReplaceCommand sends a text editor str_replace command to the R client and waits for response
func (m *WebSocketManager) sendTextEditorStrReplaceCommand(input textEditorStrReplaceInput) (*textEditorStrReplaceOutput, error) {
	// Create response channel
	responseChan := make(chan *textEditorStrReplaceOutput, 1)

	// Send the tool command
	requestID, err := m.sendToolCommand("text_editor", "str_replace", input, responseChan)
	if err != nil {
		close(responseChan)
		return nil, err
	}

	// Clean up function
	cleanup := func() {
		m.requestMu.Lock()
		delete(m.pendingRequests, requestID)
		m.requestMu.Unlock()
		close(responseChan)
	}

	// Wait for response with timeout
	select {
	case response := <-responseChan:
		cleanup()
		return response, nil
	case <-time.After(30 * time.Second):
		cleanup()
		return nil, ErrTimeout
	}
}

// sendTextEditorCreateCommand sends a text editor create command to the R client and waits for response
func (m *WebSocketManager) sendTextEditorCreateCommand(input textEditorCreateInput) (*textEditorCreateOutput, error) {
	// Create response channel
	responseChan := make(chan *textEditorCreateOutput, 1)

	// Send the tool command
	requestID, err := m.sendToolCommand("text_editor", "create", input, responseChan)
	if err != nil {
		close(responseChan)
		return nil, err
	}

	// Clean up function
	cleanup := func() {
		m.requestMu.Lock()
		delete(m.pendingRequests, requestID)
		m.requestMu.Unlock()
		close(responseChan)
	}

	// Wait for response with timeout
	select {
	case response := <-responseChan:
		cleanup()
		return response, nil
	case <-time.After(30 * time.Second):
		cleanup()
		return nil, ErrTimeout
	}
}

// sendTextEditorInsertCommand sends a text editor insert command to the R client and waits for response
func (m *WebSocketManager) sendTextEditorInsertCommand(input textEditorInsertInput) (*textEditorInsertOutput, error) {
	// Create response channel
	responseChan := make(chan *textEditorInsertOutput, 1)

	// Send the tool command
	requestID, err := m.sendToolCommand("text_editor", "insert", input, responseChan)
	if err != nil {
		close(responseChan)
		return nil, err
	}

	// Clean up function
	cleanup := func() {
		m.requestMu.Lock()
		delete(m.pendingRequests, requestID)
		m.requestMu.Unlock()
		close(responseChan)
	}

	// Wait for response with timeout
	select {
	case response := <-responseChan:
		cleanup()
		return response, nil
	case <-time.After(30 * time.Second):
		cleanup()
		return nil, ErrTimeout
	}
}

// handleToolResponse handles a response from the R client for any tool
func (m *WebSocketManager) handleToolResponse(requestID, tool, command string, result interface{}) {
	m.requestMu.RLock()
	responseChan, exists := m.pendingRequests[requestID]
	m.requestMu.RUnlock()

	if !exists {
		return
	}

	// For text_editor tool
	if tool == "text_editor" && command == "view" {
		if typedChan, ok := responseChan.(chan *textEditorViewOutput); ok {
			// Convert result to JSON bytes then unmarshal directly into target type
			jsonBytes, err := json.Marshal(result)
			if err != nil {
				typedChan <- &textEditorViewOutput{Error: "Failed to marshal response"}
				return
			}

			// Unmarshal directly into the target type
			var output textEditorViewOutput
			if err := json.Unmarshal(jsonBytes, &output); err != nil {
				typedChan <- &textEditorViewOutput{Error: "Failed to parse response"}
				return
			}

			typedChan <- &output
		}
	} else if tool == "text_editor" && command == "str_replace" {
		if typedChan, ok := responseChan.(chan *textEditorStrReplaceOutput); ok {
			// Convert result to JSON bytes then unmarshal directly into target type
			jsonBytes, err := json.Marshal(result)
			if err != nil {
				typedChan <- &textEditorStrReplaceOutput{Error: "Failed to marshal response"}
				return
			}

			// Unmarshal directly into the target type
			var output textEditorStrReplaceOutput
			if err := json.Unmarshal(jsonBytes, &output); err != nil {
				typedChan <- &textEditorStrReplaceOutput{Error: "Failed to parse response"}
				return
			}

			typedChan <- &output
		}
	} else if tool == "text_editor" && command == "create" {
		if typedChan, ok := responseChan.(chan *textEditorCreateOutput); ok {
			// Convert result to JSON bytes then unmarshal directly into target type
			jsonBytes, err := json.Marshal(result)
			if err != nil {
				typedChan <- &textEditorCreateOutput{Error: "Failed to marshal response"}
				return
			}

			// Unmarshal directly into the target type
			var output textEditorCreateOutput
			if err := json.Unmarshal(jsonBytes, &output); err != nil {
				typedChan <- &textEditorCreateOutput{Error: "Failed to parse response"}
				return
			}

			typedChan <- &output
		}
	} else if tool == "text_editor" && command == "insert" {
		if typedChan, ok := responseChan.(chan *textEditorInsertOutput); ok {
			// Convert result to JSON bytes then unmarshal directly into target type
			jsonBytes, err := json.Marshal(result)
			if err != nil {
				typedChan <- &textEditorInsertOutput{Error: "Failed to marshal response"}
				return
			}

			// Unmarshal directly into the target type
			var output textEditorInsertOutput
			if err := json.Unmarshal(jsonBytes, &output); err != nil {
				typedChan <- &textEditorInsertOutput{Error: "Failed to parse response"}
				return
			}

			typedChan <- &output
		}
	}
	// Future: Add handling for other tools here
}

// sendToolCommand sends a tool command to the R client and waits for response
func (m *WebSocketManager) sendToolCommand(tool, command string, input interface{}, responseChan interface{}) (string, error) {
	// Generate unique ID for this request
	requestID := uuid.New().String()

	// Store response channel
	m.requestMu.Lock()
	m.pendingRequests[requestID] = responseChan
	m.requestMu.Unlock()

	// Create message with general structure
	msg := WebSocketMessage{
		ID:      requestID,
		Type:    "tool_request",
		Tool:    tool,
		Command: command,
		Input:   input,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		// Clean up on error
		m.requestMu.Lock()
		delete(m.pendingRequests, requestID)
		m.requestMu.Unlock()
		return "", err
	}

	// Send to R client
	if err := m.sendToClient("r_tool_rpc", data); err != nil {
		// Clean up on error
		m.requestMu.Lock()
		delete(m.pendingRequests, requestID)
		m.requestMu.Unlock()
		return "", err
	}

	return requestID, nil
}