package api

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	Description: anthropic.String("List the contents (files and subdirectories) of the current working directory."),
	InputSchema: ListFilesToolInputSchema,
}

type ListFilesToolInput struct{}

var ListFilesToolInputSchema = GenerateSchema[ListFilesToolInput]()

type ListFilesToolResult struct {
	Objects []ListFilesToolResultObj `json:"objects"`
}

type ListFilesToolResultObj struct {
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
}

// HTTP client for making calls to RStudio frontend
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// readFileFromRStudio makes an HTTP call to the RStudio frontend to read a file
func (s *ServerClient) readFileFromRStudio(filePath string) (*ReadFileToolResult, error) {
	url := fmt.Sprintf("%s/assistant/read-file", s.rstudioURL)
	
	payload := map[string]string{
		"path": filePath,
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RStudio frontend returned status: %d", resp.StatusCode)
	}
	
	var result ReadFileToolResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &result, nil
}

// listFilesFromRStudio makes an HTTP call to the RStudio frontend to list files
func (s *ServerClient) listFilesFromRStudio() (*ListFilesToolResult, error) {
	url := fmt.Sprintf("%s/assistant/list-files", s.rstudioURL)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RStudio frontend returned status: %d", resp.StatusCode)
	}
	
	var result ListFilesToolResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &result, nil
}
