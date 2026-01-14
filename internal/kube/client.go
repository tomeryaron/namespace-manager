package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes clientset to interact with the Kubernetes API
type Client struct {
	clientset *kubernetes.Clientset // The Kubernetes API client
	config    *rest.Config           // Configuration for connecting to the cluster
}

// NewClient creates and returns a new Kubernetes client instance
// It automatically detects the cluster connection:
// 1. If running inside a pod: uses in-cluster config (service account)
// 2. If running locally: uses kubeconfig file (usually ~/.kube/config)
func NewClient() (*Client, error) {
	var config *rest.Config
	var err error

	// Try to get in-cluster config first (if running inside Kubernetes)
	config, err = rest.InClusterConfig()
	if err != nil {
		// If not in-cluster, try to use kubeconfig file (for local development)
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return nil, err
		}
	}

	// Create the clientset using the config
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		clientset: clientset,
		config:    config,
	}, nil
}