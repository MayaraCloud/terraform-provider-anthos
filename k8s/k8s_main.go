package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // This is needed for gcp auth
	"k8s.io/client-go/tools/clientcmd"
)

// Auth contains authentication info for kubernetes
type Auth struct {
	KubeConfigFile string
	KubeContext    string
}

// KubeClientSet initializes the kubernetes API client
func KubeClientSet(auth Auth) (*kubernetes.Clientset, error) {
	kubeConfig := auth.KubeConfigFile
	kubeContext := auth.KubeContext

	if home := homeDir(); home != "" && kubeConfig == "" {
		kubeConfig = filepath.Join(home, ".kube", "config")
	} else {
		return nil, fmt.Errorf("Homedir not found and no explicit config path provided")
	}

	// use the current context in kubeconfig
	config, err := clientcmd.LoadFromFile(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("Loading kube config from file: %w", err)
	}
	// TODO it would be good to set proper config overrides
	configOverrides := &clientcmd.ConfigOverrides{}
	var clientConfig clientcmd.ClientConfig
	if kubeConfig == "current" {
		clientConfig = clientcmd.NewDefaultClientConfig(*config, configOverrides)
	} else {
		clientConfig = clientcmd.NewNonInteractiveClientConfig(*config, kubeContext, configOverrides, nil)
	}
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("Getting Rest config from API config: %w", err)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// GetK8sClusterUUID returns the kube-system namespace UID
func GetK8sClusterUUID(ctx context.Context, auth Auth) (string, error) {
	kubeClient, err := KubeClientSet(auth)
	if err != nil {
		return "", fmt.Errorf("Initializing Kube clientset: %w", err)
	}
	namespaceName := "kube-system"
	namespace, err := kubeClient.CoreV1().Namespaces().Get(ctx, namespaceName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("Getting %v namespace details: %w", namespaceName, err)
	}

	return string(namespace.GetUID()), nil
}
