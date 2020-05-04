package k8s

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InstallOrUpdateGKEConnectAgent installs or update a gke-connect agent in a Kubernetes cluster
func InstallOrUpdateGKEConnectAgent(ctx context.Context, auth Auth, manifestResponse ConnectManifestResponse) error {
	kubeClient, err := KubeClientSet(auth)
	if err != nil {
		return fmt.Errorf("Initializing Kube clientset: %w", err)
	}

	for _, manifest := range manifestResponse.Manifest {
		var object objectWithMetadata
		err = yaml.Unmarshal([]byte(manifest.Manifest), object)
		if err != nil {
			return fmt.Errorf("Un-marshaling manifest %w", err)
		}
		if object.Kind == "Namespace" {
			var namespace v1.Namespace
			err = yaml.Unmarshal([]byte(manifest.Manifest), namespace)
			if err != nil {
				return fmt.Errorf("Un-marshaling namespace %w", err)
			}
			namespaces, err := kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
			nsIsPresent := func() bool {
				for _, ns := range namespaces.Items {
					if ns.ObjectMeta.Name == namespace.Name {
						return true
					}
				}
				return false
			}
			// Create namespace if not present
			if !nsIsPresent() {
				_, err = kubeClient.CoreV1().Namespaces().Create(ctx, &namespace, metav1.CreateOptions{})
				if err != nil {
					return fmt.Errorf("Creating namespace %w", err)
				}
			}
		}
	}

	return nil
}

type objectWithMetadata struct {
	Kind     string
	Metadata objectMetadata
}
type objectMetadata struct {
	Namespace string
	Name      string
	Labels    objectLabels
}
type objectLabels struct {
	Version string // This is the connect agent version for the object
}

// ConnectManifestResponse contains the connect agent manifest response
type ConnectManifestResponse struct {
	Manifest []ConnectAgentResource `json:"manifest"`
}

// ConnectAgentResource is part of GenerateConnectManifestResponse
type ConnectAgentResource struct {
	Type     ConnectAgentResourceType `json:"type"`
	Manifest string                   `json:"manifest"`
}

// ConnectAgentResourceType is part of ConnectAgentResource
type ConnectAgentResourceType struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
}
