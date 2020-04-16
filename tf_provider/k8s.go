package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

func kubeClientSet(d *schema.ResourceData) (*kubernetes.Clientset, error) {
	kubeConfig := d.Get("k8s_config_file").(string)
	kubeContext := d.Get("k8s_context").(string)

	if home := homeDir(); home != "" && kubeConfig == "" && kubeContext != "" {
		kubeConfig = filepath.Join(home, ".kube", "config")
	} else {
		return nil, fmt.Errorf("Homedir not found and no explicit config path provided")
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags(kubeContext, kubeConfig)
	if err != nil {
		return nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
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

func getK8sClusterUUID(d *schema.ResourceData) (string error) {
	kubeClient, err := kubeClientSet(d)
	if err != nil {
		return "", fmt.Errorf("Initializing Kube clientset: %w", err)
	}
	namespaceName := "kube-system"
	namespace, err := kubeClient.CoreV1().Namespaces().Get(namespaceName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("Getting %w namespace details: %w", namespaceName, err)
	}
	
	return namespace.GetUID(), nil
}