package main

import (
	"fmt"
	"os"

	"github.com/Aamod007/Kube-MCP/internal/k8s"
	"github.com/Aamod007/Kube-MCP/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Initialize Kubernetes client
	client, err := k8s.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	// Create MCP server
	s := server.NewMCPServer("KubeMCP", "1.0.0", server.WithResourceCapabilities(true, true), server.WithPromptCapabilities(true))

	// Register all tools
	tools.RegisterAll(s, client)

	// In the first session, we focus on stdio transport.
	fmt.Fprintf(os.Stderr, "Starting KubeMCP server on stdio transport...\n")
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
