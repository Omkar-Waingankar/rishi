package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

var ReadFileTool = anthropic.ToolParam{
	Name:        "read_file",
	Description: anthropic.String("Read the contents of a file. Use this when you want to see what's inside a file (usually code). Do not use this with directory names. Do not use this with data files."),
	InputSchema: ReadFileToolInputSchema,
}

type ReadFileToolInput struct {
	Path string `json:"path"`
}

type ReadFileToolResult struct {
	Content string `json:"content"`
}

var ReadFileToolInputSchema = GenerateSchema[ReadFileToolInput]()

var ListFilesTool = anthropic.ToolParam{
	Name:        "list_files",
	Description: anthropic.String("List the contents (files and subdirectories) of a directory. If no path is provided, lists the current working directory. You have the option of listing subdirectories recursively or not."),
	InputSchema: ListFilesToolInputSchema,
}

type ListFilesToolInput struct {
	Path      string `json:"path,omitempty"`
	Recursive bool   `json:"recursive,omitempty"`
}

var ListFilesToolInputSchema = GenerateSchema[ListFilesToolInput]()

type ListFilesToolResult struct {
	Objects []ListFilesToolResultObj `json:"objects"`
}

type ListFilesToolResultObj struct {
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
}

// HTTP client with timeout for RPC calls
var rpcClient = &http.Client{
	Timeout: 10 * time.Second,
}

// makeRPCRequest makes an authenticated HTTP request to the Tool RPC server
func (s *ServerClient) makeRPCRequest(method, endpoint string, payload interface{}) (*http.Response, error) {
	var bodyReader io.Reader

	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, fmt.Sprintf("http://127.0.0.1:8082%s", endpoint), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.toolRPCToken))

	resp, err := rpcClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RPC request failed: %w", err)
	}

	return resp, nil
}

// readFileFromRStudio makes an HTTP call to the Tool RPC server to read a file
func (s *ServerClient) readFileFromRStudio(filePath string) (*ReadFileToolResult, error) {
	payload := map[string]interface{}{
		"relpath":   filePath,
		"max_bytes": 2000000,
	}

	resp, err := s.makeRPCRequest("POST", "/read", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to make RPC request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if errorMsg, ok := errorResp["error"].(string); ok {
				return &ReadFileToolResult{
					Content: fmt.Sprintf("Error: %s", errorMsg),
				}, nil
			}
		}
		return &ReadFileToolResult{
			Content: fmt.Sprintf("Error: HTTP %d", resp.StatusCode),
		}, nil
	}

	var rpcResp map[string]interface{}
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse RPC response: %w", err)
	}

	contentValue, exists := rpcResp["content"]
	if !exists {
		return nil, fmt.Errorf("invalid response format: missing content field")
	}

	content, ok := contentValue.(string)
	if !ok {
		return nil, fmt.Errorf("invalid response format: content is not a string")
	}

	return &ReadFileToolResult{
		Content: content,
	}, nil
}

// listFilesFromRStudio makes an HTTP call to the Tool RPC server to list files
func (s *ServerClient) listFilesFromRStudio(path string, recursive bool) (*ListFilesToolResult, error) {
	payload := map[string]interface{}{
		"path":      path,
		"pattern":   nil,
		"recursive": recursive,
		"max_items": 50,
	}

	resp, err := s.makeRPCRequest("POST", "/list", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to make RPC request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if errorMsg, ok := errorResp["error"].(string); ok {
				// Return empty list with error in a file entry for visibility
				return &ListFilesToolResult{
					Objects: []ListFilesToolResultObj{
						{Path: fmt.Sprintf("Error: %s", errorMsg), IsDir: false},
					},
				}, nil
			}
		}
		return &ListFilesToolResult{
			Objects: []ListFilesToolResultObj{
				{Path: fmt.Sprintf("Error: HTTP %d", resp.StatusCode), IsDir: false},
			},
		}, nil
	}

	var rpcResp map[string]interface{}
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse RPC response: %w", err)
	}

	filesInterface, ok := rpcResp["files"]
	if !ok {
		return nil, fmt.Errorf("invalid response format: missing files field")
	}

	files, ok := filesInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: files is not an array")
	}

	var objects []ListFilesToolResultObj
	for _, fileInterface := range files {
		filename, ok := fileInterface.(string)
		if !ok {
			continue
		}

		// Simple heuristic: assume it's a directory if it doesn't have an extension
		isDir := len(filename) > 0 && filename[len(filename)-1] == '/' ||
			(len(filename) > 0 && !containsDot(filename))

		objects = append(objects, ListFilesToolResultObj{
			Path:  filename,
			IsDir: isDir,
		})
	}

	return &ListFilesToolResult{
		Objects: objects,
	}, nil
}

// containsDot checks if a filename contains a dot (indicating it likely has an extension)
func containsDot(filename string) bool {
	for _, char := range filename {
		if char == '.' {
			return true
		}
	}
	return false
}
