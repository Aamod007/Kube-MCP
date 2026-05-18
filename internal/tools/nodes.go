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

type nodeSummary struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Roles   string `json:"roles"`
	Version string `json:"version"`
	OS      string `json:"os"`
	Age     string `json:"age"`
}

func registerNodeTools(s *server.MCPServer, client *k8s.Client) {
	// list_nodes
	listNodesTool := mcp.NewTool("list_nodes",
		mcp.WithDescription("All nodes with status, roles, kernel, runtime"),
	)
	s.AddTool(listNodesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		nodes, err := client.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return createErrorResponse(fmt.Errorf("failed to list nodes: %w", err)), nil
		}

		var summaries []nodeSummary
		for _, node := range nodes.Items {
			status := "NotReady"
			for _, condition := range node.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == "True" {
					status = "Ready"
					break
				}
			}

			var roles []string
			for label := range node.Labels {
				if strings.HasPrefix(label, "node-role.kubernetes.io/") {
					roles = append(roles, strings.TrimPrefix(label, "node-role.kubernetes.io/"))
				}
			}
			roleStr := "<none>"
			if len(roles) > 0 {
				roleStr = strings.Join(roles, ",")
			}

			summaries = append(summaries, nodeSummary{
				Name:    node.Name,
				Status:  status,
				Roles:   roleStr,
				Version: node.Status.NodeInfo.KubeletVersion,
				OS:      node.Status.NodeInfo.OSImage,
			})
		}

		return createTextResponse(formatJSON(summaries)), nil
	})

	// describe_node
	describeNodeTool := mcp.NewTool("describe_node",
		mcp.WithDescription("Node capacity, allocatable, conditions, taints"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Node name")),
	)
	s.AddTool(describeNodeTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return createErrorResponse(fmt.Errorf("invalid arguments")), nil
		}
		name, ok := args["name"].(string)
		if !ok {
			return createErrorResponse(fmt.Errorf("name argument is required")), nil
		}

		node, err := client.Clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return createErrorResponse(fmt.Errorf("failed to get node: %w", err)), nil
		}

		// Provide a truncated but useful summary
		summary := map[string]interface{}{
			"name":        node.Name,
			"labels":      node.Labels,
			"annotations": node.Annotations,
			"taints":      node.Spec.Taints,
			"capacity":    node.Status.Capacity,
			"allocatable": node.Status.Allocatable,
			"conditions":  node.Status.Conditions,
			"info":        node.Status.NodeInfo,
		}

		return createTextResponse(formatJSON(summary)), nil
	})
}
