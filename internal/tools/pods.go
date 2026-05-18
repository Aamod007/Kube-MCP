package tools

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Aamod007/Kube-MCP/internal/k8s"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type podSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Restarts  int32  `json:"restarts"`
	Ready     string `json:"ready"`
	Node      string `json:"node"`
	Age       string `json:"age"`
}

func registerPodTools(s *server.MCPServer, client *k8s.Client) {
	// list_pods
	listPodsTool := mcp.NewTool("list_pods",
		mcp.WithDescription("List pods with status, restarts, age, and node. Returns a summary of pods in the specified namespace."),
		mcp.WithString("namespace", mcp.Description("Kubernetes namespace (default: default)"), mcp.DefaultString("default")),
	)
	s.AddTool(listPodsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := "default"
		args, ok := request.Params.Arguments.(map[string]interface{})
		if ok {
			if ns, ok := args["namespace"].(string); ok && ns != "" {
				namespace = ns
			}
		}

		pods, err := client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return createErrorResponse(fmt.Errorf("failed to list pods: %w", err)), nil
		}

		var summaries []podSummary
		for _, pod := range pods.Items {
			restarts, readyContainers, totalContainers := getPodContainerStats(&pod)
			age := time.Since(pod.CreationTimestamp.Time).Round(time.Second).String()

			summaries = append(summaries, podSummary{
				Name:      pod.Name,
				Namespace: pod.Namespace,
				Status:    string(pod.Status.Phase),
				Restarts:  restarts,
				Ready:     fmt.Sprintf("%d/%d", readyContainers, totalContainers),
				Node:      pod.Spec.NodeName,
				Age:       age,
			})
		}

		return createTextResponse(formatJSON(summaries)), nil
	})

	// get_pod
	getPodTool := mcp.NewTool("get_pod",
		mcp.WithDescription("Fetch detailed pod spec and status for a specific pod."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Pod name")),
		mcp.WithString("namespace", mcp.Description("Kubernetes namespace (default: default)"), mcp.DefaultString("default")),
	)
	s.AddTool(getPodTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return createErrorResponse(fmt.Errorf("invalid arguments")), nil
		}
		name, ok := args["name"].(string)
		if !ok {
			return createErrorResponse(fmt.Errorf("name argument is required")), nil
		}

		namespace := "default"
		if ns, ok := args["namespace"].(string); ok && ns != "" {
			namespace = ns
		}

		pod, err := client.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return createErrorResponse(fmt.Errorf("failed to get pod: %w", err)), nil
		}

		return createTextResponse(formatJSON(pod)), nil
	})

	// get_logs
	getLogsTool := mcp.NewTool("get_logs",
		mcp.WithDescription("Fetch logs from a container in a pod. Returns the last N lines."),
		mcp.WithString("pod", mcp.Required(), mcp.Description("Pod name")),
		mcp.WithString("namespace", mcp.Description("Kubernetes namespace (default: default)"), mcp.DefaultString("default")),
		mcp.WithString("container", mcp.Description("Container name (omit for single-container pods)")),
		mcp.WithNumber("tail", mcp.Description("Number of lines from end"), mcp.DefaultNumber(100)),
		mcp.WithString("previous", mcp.Description("Set to 'true' to fetch logs from previous container instance (useful for crashloops)")),
	)
	s.AddTool(getLogsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return createErrorResponse(fmt.Errorf("invalid arguments")), nil
		}
		podName, ok := args["pod"].(string)
		if !ok {
			return createErrorResponse(fmt.Errorf("pod argument is required")), nil
		}

		namespace := "default"
		if ns, ok := args["namespace"].(string); ok && ns != "" {
			namespace = ns
		}

		logOptions := &corev1.PodLogOptions{}

		if container, ok := args["container"].(string); ok && container != "" {
			logOptions.Container = container
		}

		if tailVal, ok := args["tail"].(float64); ok {
			tail := int64(tailVal)
			logOptions.TailLines = &tail
		} else {
			tail := int64(100)
			logOptions.TailLines = &tail
		}

		if prev, ok := args["previous"].(string); ok && prev == "true" {
			logOptions.Previous = true
		}

		req := client.Clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
		podLogs, err := req.Stream(ctx)
		if err != nil {
			return createErrorResponse(fmt.Errorf("failed to open log stream: %w", err)), nil
		}
		defer podLogs.Close()

		// Read up to 50KB to avoid massive token usage
		buf := make([]byte, 50*1024)
		n, err := io.ReadFull(podLogs, buf)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return createErrorResponse(fmt.Errorf("error reading logs: %w", err)), nil
		}

		logs := string(buf[:n])
		if n == len(buf) {
			logs += "\n[truncated...]"
		}

		return createTextResponse(logs), nil
	})
}

func getPodContainerStats(pod *corev1.Pod) (restarts int32, readyContainers int, totalContainers int) {
	totalContainers = len(pod.Spec.Containers)
	for _, cs := range pod.Status.ContainerStatuses {
		restarts += cs.RestartCount
		if cs.Ready {
			readyContainers++
		}
	}
	return
}
