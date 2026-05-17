package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func (c *Client) GetNodeMetrics(ctx context.Context, name string) (*v1beta1.NodeMetrics, error) {
	if c.MetricsClientset == nil {
		return nil, fmt.Errorf("metrics-server is not available in the cluster")
	}
	return c.MetricsClientset.MetricsV1beta1().NodeMetricses().Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) ListPodMetrics(ctx context.Context, namespace string) (*v1beta1.PodMetricsList, error) {
	if c.MetricsClientset == nil {
		return nil, fmt.Errorf("metrics-server is not available in the cluster")
	}
	return c.MetricsClientset.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
}

func (c *Client) GetPodMetrics(ctx context.Context, namespace, name string) (*v1beta1.PodMetrics, error) {
	if c.MetricsClientset == nil {
		return nil, fmt.Errorf("metrics-server is not available in the cluster")
	}
	return c.MetricsClientset.MetricsV1beta1().PodMetricses(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) ListNodeMetrics(ctx context.Context) (*v1beta1.NodeMetricsList, error) {
	if c.MetricsClientset == nil {
		return nil, fmt.Errorf("metrics-server is not available in the cluster")
	}
	return c.MetricsClientset.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
}
