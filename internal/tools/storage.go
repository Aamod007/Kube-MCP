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

type pvcSummary struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	Status       string `json:"status"`
	Volume       string `json:"volume"`
	Capacity     string `json:"capacity"`
	StorageClass string `json:"storageClass"`
	Age          string `json:"age"`
}

func registerStorageTools(s *server.MCPServer, client *k8s.Client) {
	// get_pvc_status
	getPvcStatusTool := mcp.NewTool("get_pvc_status",
		mcp.WithDescription("PersistentVolumeClaim binding and capacity"),
		mcp.WithString("namespace", mcp.Description("Kubernetes namespace (default: default)"), mcp.DefaultString("default")),
	)
	s.AddTool(getPvcStatusTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := "default"
		args, ok := request.Params.Arguments.(map[string]interface{})
		if ok {
			if ns, ok := args["namespace"].(string); ok && ns != "" {
				namespace = ns
			}
		}

		pvcs, err := client.Clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return createErrorResponse(fmt.Errorf("failed to list PVCs: %w", err)), nil
		}

		var summaries []pvcSummary
		for _, pvc := range pvcs.Items {
			capacity := ""
			if q, ok := pvc.Status.Capacity["storage"]; ok {
				capacity = q.String()
			}
			sc := ""
			if pvc.Spec.StorageClassName != nil {
				sc = *pvc.Spec.StorageClassName
			}
			age := time.Since(pvc.CreationTimestamp.Time).Round(time.Second).String()

			summaries = append(summaries, pvcSummary{
				Name:         pvc.Name,
				Namespace:    pvc.Namespace,
				Status:       string(pvc.Status.Phase),
				Volume:       pvc.Spec.VolumeName,
				Capacity:     capacity,
				StorageClass: sc,
				Age:          age,
			})
		}

		return createTextResponse(formatJSON(summaries)), nil
	})
}
