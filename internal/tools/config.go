package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/Aamod007/Kube-MCP/internal/k8s"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type configMapSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Keys      string `json:"keys"`
}

func registerConfigTools(s *server.MCPServer, client *k8s.Client) {
	// list_configmaps
	listConfigMapsTool := mcp.NewTool("list_configmaps",
		mcp.WithDescription("ConfigMap names and keys (not values by default)"),
		mcp.WithString("namespace", mcp.Description("Kubernetes namespace (default: default)"), mcp.DefaultString("default")),
	)
	s.AddTool(listConfigMapsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := "default"
		args, ok := request.Params.Arguments.(map[string]interface{})
		if ok {
			if ns, ok := args["namespace"].(string); ok && ns != "" {
				namespace = ns
			}
		}

		cms, err := client.Clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return createErrorResponse(fmt.Errorf("failed to list configmaps: %w", err)), nil
		}

		var summaries []configMapSummary
		for _, cm := range cms.Items {
			var keys []string
			for k := range cm.Data {
				keys = append(keys, k)
			}
			summaries = append(summaries, configMapSummary{
				Name:      cm.Name,
				Namespace: cm.Namespace,
				Keys:      strings.Join(keys, ","),
			})
		}

		return createTextResponse(formatJSON(summaries)), nil
	})
}
