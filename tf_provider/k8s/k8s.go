package k8s

import (
	"fmt"
	"strings"
	"os"
	"context"
	"path/filepath"
	"github.com/ghodss/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // This is needed for gcp auth
    "gitlab.com/mayara/private/anthos/debug"
)

// Auth contains authentication info for kubernetes
type Auth struct {
	KubeConfigFile string
	KubeContext string
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

// Absolute Kubernetes API paths of the exclusivity artifacts
const (
	CRDAbspath string = "apis/apiextensions.k8s.io/v1beta1/customresourcedefinitions/memberships.hub.gke.io"
	CRAbspath = "apis/hub.gke.io/v1/memberships/membership"
)

// GetMembershipCR get the Membership CR
func GetMembershipCR(ctx context.Context, auth Auth) (string, error) {
	kubeClient, err := KubeClientSet(auth)
	if err != nil {
		return "", fmt.Errorf("Initializing Kube clientset: %w", err)
	}
	object, err := kubeClient.RESTClient().Get().AbsPath(CRAbspath).DoRaw(ctx)
	if err != nil {
		return "", fmt.Errorf("Getting the membership object: %w", err)
	}
	yamlObject, err := yaml.JSONToYAML(object)
	if err != nil {
		return "", fmt.Errorf("Transforming CR Manifest json to yaml: %w", err)
	}
	return string(string(yamlObject)), nil
}

// GetMembershipCRD get the Membership CRD
func GetMembershipCRD(ctx context.Context, auth Auth) (string, error) {
	kubeClient, err := KubeClientSet(auth)
	if err != nil {
		return "", fmt.Errorf("Initializing Kube clientset: %w", err)
	}
	object, err := kubeClient.RESTClient().Get().AbsPath(CRDAbspath).DoRaw(ctx)
	if err != nil {
		// If there is no Membership CRD we just return an empty string
		if strings.Contains(err.Error(), "the server could not find the requested resource") {		
			return "", nil
		}
		return "", fmt.Errorf("Getting the membership CRD object: %w", err)
	}
	yamlObject, err := yaml.JSONToYAML(object)
	if err != nil {
		return "", fmt.Errorf("Transforming CRD Manifest json to yaml: %w", err)
	}
	return string(string(yamlObject)), nil
}

// InstallExclusivityManifests applies the CRD and CR manifests in the cluster
// This will either install or upgrade them if already present
func InstallExclusivityManifests(ctx context.Context, auth Auth, CRDManifest string, CRManifest string) error {
	kubeClient, err := KubeClientSet(auth)
	if err != nil {
		return fmt.Errorf("Initializing Kube clientset: %w", err)
	}
	if CRDManifest != "" {
		debug.GoLog("InstallExclusivityManifests: installing CRD manifest")
		err = installRawArtifact(ctx, kubeClient, CRDAbspath, CRDManifest)
		if err != nil {
			return fmt.Errorf("Installing CRD: %w", err)
		}
	}
	if CRManifest != "" {
		debug.GoLog("InstallExclusivityManifests: installing CR manifest")
		err = installRawArtifact(ctx, kubeClient, CRAbspath, CRManifest)
		if err != nil {
			return fmt.Errorf("Installing CR: %w", err)
		}
	}

	debug.GoLog("InstallExclusivityManifests: received empty artifacts")
	return nil
}

func installRawArtifact(ctx context.Context, kubeClient *kubernetes.Clientset, absPath string, artifact string) error {

	_, err := kubeClient.RESTClient().Get().AbsPath(absPath).DoRaw(ctx)
	if err != nil {
		// If there is no artifact CREATE, otherwise, PATCH
		if strings.Contains(err.Error(), "the server could not find the requested resource") {		
			debug.GoLog("installRawArtifact: installing the artifact " + absPath)
			_, err = kubeClient.RESTClient().Post().Body([]byte(artifact)).AbsPath(absPath).DoRaw(ctx)
			if err != nil {
				return fmt.Errorf("Error CREATING %v: %w", absPath, err)
			}
		}
		return fmt.Errorf("Getting %v: %w", absPath, err)
	}
	
	debug.GoLog("installRawArtifact: updating the artifact " + absPath)
	_, err = kubeClient.RESTClient().Patch(k8sTypes.ApplyPatchType).Body([]byte(artifact)).AbsPath(absPath).DoRaw(ctx)
	if err != nil {
		return fmt.Errorf("Error PATCHING %v: %w", absPath, err)
	}

	return nil
}

// DeleteArtifacts deletes the CRD and CR manifests in the cluster
func DeleteArtifacts(ctx context.Context, auth Auth) error {
	kubeClient, err := KubeClientSet(auth)
	if err != nil {
		return fmt.Errorf("Initializing Kube clientset: %w", err)
	}
	artifacts := []string{ CRDAbspath, CRAbspath }
	for _, artifact := range(artifacts) {
		// Check if the artifact exists
		_, err := kubeClient.RESTClient().Get().AbsPath(artifact).DoRaw(ctx)
		if err != nil {
			// If there is no artifact no need to delete it
			if strings.Contains(err.Error(), "the server could not find the requested resource") {		
				return nil
			}
			return fmt.Errorf("Getting the artifact before deleting: %w", err)
		}
		// Delete the resource
		_, err = kubeClient.RESTClient().Delete().AbsPath(artifact).DoRaw(ctx)
		if err != nil {
			return fmt.Errorf("Error DELETING %v: %w", artifact, err)
		}
	}
	
	return nil
}

