package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Client encapsulates the standard Kubernetes clientset and metrics clientset.
type Client struct {
	Clientset        *kubernetes.Clientset
	MetricsClientset *metrics.Clientset
}

// NewClient automatically detects the environment and returns a new Client.
func NewClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig
		kubeconfigPath := os.Getenv("KUBECONFIG")
		if kubeconfigPath == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeconfigPath = filepath.Join(home, ".kube", "config")
			} else {
				return nil, fmt.Errorf("could not find kubeconfig path")
			}
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	metricsClientset, err := metrics.NewForConfig(config)
	if err != nil {
		// We don't fail hard here, metrics might not be available, but we store nil
		metricsClientset = nil
	}

	return &Client{
		Clientset:        clientset,
		MetricsClientset: metricsClientset,
	}, nil
}
