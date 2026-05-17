package tools

import (
	"github.com/Aamod007/Kube-MCP/internal/k8s"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterAll registers all KubeMCP tools with the provided mcp-go server.
func RegisterAll(s *server.MCPServer, client *k8s.Client) {
	// These will be implemented in the respective files
	registerPodTools(s, client)
	registerNodeTools(s, client)
	registerEventTools(s, client)
	registerDeploymentTools(s, client)
	registerServiceTools(s, client)
	registerConfigTools(s, client)
	registerStorageTools(s, client)
	registerNamespaceTools(s, client)
	registerResourceUsageTools(s, client)
}
