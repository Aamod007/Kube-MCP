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

type eventSummary struct {
	Type    string `json:"type"`
	Reason  string `json:"reason"`
	Age     string `json:"age"`
	From    string `json:"from"`
	Message string `json:"message"`
	Object  string `json:"object"`
}

func registerEventTools(s *server.MCPServer, client *k8s.Client) {
	// get_events
	getEventsTool := mcp.NewTool("get_events",
		mcp.WithDescription("Namespace-scoped events, filterable by reason"),
		mcp.WithString("namespace", mcp.Description("Kubernetes namespace (default: default)"), mcp.DefaultString("default")),
		mcp.WithString("reason", mcp.Description("Optional filter by reason (e.g., Failed, Evicted)")),
	)
	s.AddTool(getEventsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := "default"
		args, ok := request.Params.Arguments.(map[string]interface{})
		if ok {
			if ns, ok := args["namespace"].(string); ok && ns != "" {
				namespace = ns
			}
		}

		filterReason := ""
		if ok {
			if r, ok := args["reason"].(string); ok {
				filterReason = r
			}
		}

		events, err := client.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return createErrorResponse(fmt.Errorf("failed to list events: %w", err)), nil
		}

		var summaries []eventSummary
		for _, e := range events.Items {
			if filterReason != "" && e.Reason != filterReason {
				continue
			}
			age := time.Since(e.LastTimestamp.Time).Round(time.Second).String()
			if e.LastTimestamp.IsZero() {
				age = "<unknown>"
			}
			summaries = append(summaries, eventSummary{
				Type:    e.Type,
				Reason:  e.Reason,
				Age:     age,
				From:    e.Source.Component,
				Message: e.Message,
				Object:  fmt.Sprintf("%s/%s", e.InvolvedObject.Kind, e.InvolvedObject.Name),
			})
		}

		return createTextResponse(formatJSON(summaries)), nil
	})
}
