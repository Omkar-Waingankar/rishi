package api

import (
	"fmt"

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

// readFileFromRStudio makes an HTTP call to the RStudio frontend to read a file
func (s *ServerClient) readFileFromRStudio(filePath string) (*ReadFileToolResult, error) {
	return &ReadFileToolResult{
		Content: fmt.Sprintf("This is a placeholder for the read_file tools, filePath: %s", filePath),
	}, nil
}

// listFilesFromRStudio makes an HTTP call to the RStudio frontend to list files
func (s *ServerClient) listFilesFromRStudio() (*ListFilesToolResult, error) {
	return &ListFilesToolResult{
		Objects: []ListFilesToolResultObj{
			{Path: "test.txt", IsDir: false},
			{Path: "test.csv", IsDir: false},
			{Path: "test.R", IsDir: false},
		},
	}, nil
}
