package k8s

import (
	"fmt"
	"context"
)


// InstallOrUpdateGKEConnectAgent installs or update a gke-connect agent in a Kubernetes cluster
func InstallOrUpdateGKEConnectAgent(ctx context.Context, auth Auth) error {
	kubeClient, err := KubeClientSet(auth)
	if err != nil {
		return fmt.Errorf("Initializing Kube clientset: %w", err)
	}
	_ = kubeClient
	
	return nil
}