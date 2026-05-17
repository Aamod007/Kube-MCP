package tools

import (
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
)

// formatJSON encodes an object to JSON string. Returns an error message if it fails.
func formatJSON(obj interface{}) string {
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return `{"error": "failed to encode output"}`
	}
	return string(b)
}

// createTextResponse wraps a plain string response in the format mcp-go expects.
func createTextResponse(text string) *mcp.CallToolResult {
	return mcp.NewToolResultText(text)
}

// createErrorResponse creates an error result with text.
func createErrorResponse(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(err.Error())
}
