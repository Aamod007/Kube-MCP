package tools

import (
	"context"
	"fmt"

	"github.com/Aamod007/Kube-MCP/internal/k8s"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deploymentSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Replicas  string `json:"replicas"`
	Available string `json:"available"`
	Image     string `json:"image"`
}

type rolloutSummary struct {
	Deployment string `json:"deployment"`
	Namespace  string `json:"namespace"`
	Status     string `json:"status"`
	Desired    int32  `json:"desired"`
	Updated    int32  `json:"updated"`
	Ready      int32  `json:"ready"`
	Available  int32  `json:"available"`
	Message    string `json:"message"`
}

func registerDeploymentTools(s *server.MCPServer, client *k8s.Client) {
	// list_deployments
	listDeploymentsTool := mcp.NewTool("list_deployments",
		mcp.WithDescription("Deployments with replica status and image"),
		mcp.WithString("namespace", mcp.Description("Kubernetes namespace (default: default)"), mcp.DefaultString("default")),
	)
	s.AddTool(listDeploymentsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := "default"
		args, ok := request.Params.Arguments.(map[string]interface{})
		if ok {
			if ns, ok := args["namespace"].(string); ok && ns != "" {
				namespace = ns
			}
		}

		deps, err := client.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return createErrorResponse(fmt.Errorf("failed to list deployments: %w", err)), nil
		}

		var summaries []deploymentSummary
		for _, dep := range deps.Items {
			image := "<unknown>"
			if len(dep.Spec.Template.Spec.Containers) > 0 {
				image = dep.Spec.Template.Spec.Containers[0].Image
			}
			summaries = append(summaries, deploymentSummary{
				Name:      dep.Name,
				Namespace: dep.Namespace,
				Replicas:  fmt.Sprintf("%d/%d", dep.Status.ReadyReplicas, *dep.Spec.Replicas),
				Available: fmt.Sprintf("%d", dep.Status.AvailableReplicas),
				Image:     image,
			})
		}

		return createTextResponse(formatJSON(summaries)), nil
	})

	// get_deployment
	getDeploymentTool := mcp.NewTool("get_deployment",
		mcp.WithDescription("Full deployment spec including strategy"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Deployment name")),
		mcp.WithString("namespace", mcp.Description("Kubernetes namespace (default: default)"), mcp.DefaultString("default")),
	)
	s.AddTool(getDeploymentTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		dep, err := client.Clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return createErrorResponse(fmt.Errorf("failed to get deployment: %w", err)), nil
		}

		return createTextResponse(formatJSON(dep)), nil
	})

	// check_rollout_status
	checkRolloutStatusTool := mcp.NewTool("check_rollout_status",
		mcp.WithDescription("Deployment/StatefulSet rollout progress snapshot"),
		mcp.WithString("deployment", mcp.Required(), mcp.Description("Deployment name")),
		mcp.WithString("namespace", mcp.Description("Kubernetes namespace (default: default)"), mcp.DefaultString("default")),
	)
	s.AddTool(checkRolloutStatusTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return createErrorResponse(fmt.Errorf("invalid arguments")), nil
		}
		name, ok := args["deployment"].(string)
		if !ok {
			return createErrorResponse(fmt.Errorf("deployment argument is required")), nil
		}

		namespace := "default"
		if ns, ok := args["namespace"].(string); ok && ns != "" {
			namespace = ns
		}

		dep, err := client.Clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return createErrorResponse(fmt.Errorf("failed to get deployment: %w", err)), nil
		}

		desired := int32(0)
		if dep.Spec.Replicas != nil {
			desired = *dep.Spec.Replicas
		}

		status := "in_progress"
		message := fmt.Sprintf("Waiting for %d of %d updated replicas to become available", dep.Status.UpdatedReplicas, desired)

		if dep.Generation <= dep.Status.ObservedGeneration {
			if dep.Status.UpdatedReplicas == desired && dep.Status.Replicas == desired && dep.Status.AvailableReplicas == desired {
				status = "complete"
				message = "Rollout complete"
			}
		} else {
			message = "Waiting for deployment spec update to be observed..."
		}

		summary := rolloutSummary{
			Deployment: dep.Name,
			Namespace:  dep.Namespace,
			Status:     status,
			Desired:    desired,
			Updated:    dep.Status.UpdatedReplicas,
			Ready:      dep.Status.ReadyReplicas,
			Available:  dep.Status.AvailableReplicas,
			Message:    message,
		}

		return createTextResponse(formatJSON(summary)), nil
	})
}
