package tools

import (
	"context"
	"fmt"
	"sort"

	"github.com/Aamod007/Kube-MCP/internal/k8s"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type resourceSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	CPU       string `json:"cpu"`
	Memory    string `json:"memory"`
}

func registerResourceUsageTools(s *server.MCPServer, client *k8s.Client) {
	// get_resource_usage
	getResourceUsageTool := mcp.NewTool("get_resource_usage",
		mcp.WithDescription("CPU/memory usage via metrics-server. Provide a pod_name and namespace for a specific pod, or a node_name for a specific node. Provide nothing for a cluster-wide top summary."),
		mcp.WithString("pod_name", mcp.Description("Optional specific pod name")),
		mcp.WithString("namespace", mcp.Description("Optional namespace for the pod")),
		mcp.WithString("node_name", mcp.Description("Optional specific node name")),
	)
	s.AddTool(getResourceUsageTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if client.MetricsClientset == nil {
			return createErrorResponse(fmt.Errorf("metrics-server is not available in the cluster")), nil
		}

		args, _ := request.Params.Arguments.(map[string]interface{})

		podName, _ := args["pod_name"].(string)
		namespace, _ := args["namespace"].(string)
		nodeName, _ := args["node_name"].(string)

		if podName != "" {
			if namespace == "" {
				namespace = "default"
			}
			pm, err := client.GetPodMetrics(ctx, namespace, podName)
			if err != nil {
				return createErrorResponse(fmt.Errorf("failed to get pod metrics: %w", err)), nil
			}

			cpuSum := int64(0)
			memSum := int64(0)
			for _, c := range pm.Containers {
				cpuSum += c.Usage.Cpu().MilliValue()
				memSum += c.Usage.Memory().Value()
			}

			return createTextResponse(formatJSON(resourceSummary{
				Name:      pm.Name,
				Namespace: pm.Namespace,
				CPU:       fmt.Sprintf("%dm", cpuSum),
				Memory:    fmt.Sprintf("%dMi", memSum/(1024*1024)),
			})), nil
		}

		if nodeName != "" {
			nm, err := client.GetNodeMetrics(ctx, nodeName)
			if err != nil {
				return createErrorResponse(fmt.Errorf("failed to get node metrics: %w", err)), nil
			}
			return createTextResponse(formatJSON(resourceSummary{
				Name:   nm.Name,
				CPU:    fmt.Sprintf("%dm", nm.Usage.Cpu().MilliValue()),
				Memory: fmt.Sprintf("%dMi", nm.Usage.Memory().Value()/(1024*1024)),
			})), nil
		}

		// Cluster-wide summary
		nodeMetrics, err := client.ListNodeMetrics(ctx)
		if err != nil {
			return createErrorResponse(fmt.Errorf("failed to list node metrics: %w", err)), nil
		}

		podMetrics, err := client.ListPodMetrics(ctx, "")
		if err != nil {
			return createErrorResponse(fmt.Errorf("failed to list pod metrics: %w", err)), nil
		}

		type nodeMetricData struct {
			Name     string
			CpuMilli int64
			MemBytes int64
		}
		var nodes []nodeMetricData
		for _, nm := range nodeMetrics.Items {
			nodes = append(nodes, nodeMetricData{
				Name:     nm.Name,
				CpuMilli: nm.Usage.Cpu().MilliValue(),
				MemBytes: nm.Usage.Memory().Value(),
			})
		}
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].CpuMilli > nodes[j].CpuMilli
		})

		type podMetricData struct {
			Name      string
			Namespace string
			CpuMilli  int64
			MemBytes  int64
		}
		var pods []podMetricData
		for _, pm := range podMetrics.Items {
			cpuSum := int64(0)
			memSum := int64(0)
			for _, c := range pm.Containers {
				cpuSum += c.Usage.Cpu().MilliValue()
				memSum += c.Usage.Memory().Value()
			}
			pods = append(pods, podMetricData{
				Name:      pm.Name,
				Namespace: pm.Namespace,
				CpuMilli:  cpuSum,
				MemBytes:  memSum,
			})
		}
		sort.Slice(pods, func(i, j int) bool {
			return pods[i].MemBytes > pods[j].MemBytes
		})

		var topNodes []resourceSummary
		for i := 0; i < len(nodes) && i < 5; i++ {
			topNodes = append(topNodes, resourceSummary{
				Name:   nodes[i].Name,
				CPU:    fmt.Sprintf("%dm", nodes[i].CpuMilli),
				Memory: fmt.Sprintf("%dMi", nodes[i].MemBytes/(1024*1024)),
			})
		}

		var topPods []resourceSummary
		for i := 0; i < len(pods) && i < 5; i++ {
			topPods = append(topPods, resourceSummary{
				Name:      pods[i].Name,
				Namespace: pods[i].Namespace,
				CPU:       fmt.Sprintf("%dm", pods[i].CpuMilli),
				Memory:    fmt.Sprintf("%dMi", pods[i].MemBytes/(1024*1024)),
			})
		}

		summary := map[string]interface{}{
			"top_nodes_by_cpu":   topNodes,
			"top_pods_by_memory": topPods,
		}

		return createTextResponse(formatJSON(summary)), nil
	})
}
