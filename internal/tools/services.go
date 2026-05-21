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

type serviceSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	ClusterIP string `json:"clusterIP"`
	Ports     string `json:"ports"`
}

type ingressSummary struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Hosts     string `json:"hosts"`
	Address   string `json:"address"`
}

func registerServiceTools(s *server.MCPServer, client *k8s.Client) {
	// list_services
	listServicesTool := mcp.NewTool("list_services",
		mcp.WithDescription("Services with type, ClusterIP, ports"),
		mcp.WithString("namespace", mcp.Description("Kubernetes namespace (default: default)"), mcp.DefaultString("default")),
	)
	s.AddTool(listServicesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := "default"
		args, ok := request.Params.Arguments.(map[string]interface{})
		if ok {
			if ns, ok := args["namespace"].(string); ok && ns != "" {
				namespace = ns
			}
		}

		svcs, err := client.Clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return createErrorResponse(fmt.Errorf("failed to list services: %w", err)), nil
		}

		var summaries []serviceSummary
		for _, svc := range svcs.Items {
			var ports []string
			for _, p := range svc.Spec.Ports {
				ports = append(ports, fmt.Sprintf("%d/%s", p.Port, p.Protocol))
			}
			summaries = append(summaries, serviceSummary{
				Name:      svc.Name,
				Namespace: svc.Namespace,
				Type:      string(svc.Spec.Type),
				ClusterIP: svc.Spec.ClusterIP,
				Ports:     strings.Join(ports, ","),
			})
		}

		return createTextResponse(formatJSON(summaries)), nil
	})

	// list_ingresses
	listIngressesTool := mcp.NewTool("list_ingresses",
		mcp.WithDescription("Ingress rules, hosts, TLS status"),
		mcp.WithString("namespace", mcp.Description("Kubernetes namespace (default: default)"), mcp.DefaultString("default")),
	)
	s.AddTool(listIngressesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := "default"
		args, ok := request.Params.Arguments.(map[string]interface{})
		if ok {
			if ns, ok := args["namespace"].(string); ok && ns != "" {
				namespace = ns
			}
		}

		ings, err := client.Clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return createErrorResponse(fmt.Errorf("failed to list ingresses: %w", err)), nil
		}

		var summaries []ingressSummary
		for _, ing := range ings.Items {
			var hosts []string
			for _, rule := range ing.Spec.Rules {
				hosts = append(hosts, rule.Host)
			}

			var addresses []string
			for _, lb := range ing.Status.LoadBalancer.Ingress {
				if lb.IP != "" {
					addresses = append(addresses, lb.IP)
				}
				if lb.Hostname != "" {
					addresses = append(addresses, lb.Hostname)
				}
			}

			summaries = append(summaries, ingressSummary{
				Name:      ing.Name,
				Namespace: ing.Namespace,
				Hosts:     strings.Join(hosts, ","),
				Address:   strings.Join(addresses, ","),
			})
		}

		return createTextResponse(formatJSON(summaries)), nil
	})
}
