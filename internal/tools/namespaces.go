package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/Aamod007/Kube-MCP/internal/k8s"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type namespaceSummary struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Age    string `json:"age"`
}

func registerNamespaceTools(s *server.MCPServer, client *k8s.Client) {
	// list_namespaces
	listNamespacesTool := mcp.NewTool("list_namespaces",
		mcp.WithDescription("All namespaces with status and labels"),
	)
	s.AddTool(listNamespacesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		nsList, err := client.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return createErrorResponse(fmt.Errorf("failed to list namespaces: %w", err)), nil
		}

		var summaries []namespaceSummary
		for _, ns := range nsList.Items {
			age := time.Since(ns.CreationTimestamp.Time).Round(time.Second).String()
			summaries = append(summaries, namespaceSummary{
				Name:   ns.Name,
				Status: string(ns.Status.Phase),
				Age:    age,
			})
		}

		return createTextResponse(formatJSON(summaries)), nil
	})
}
