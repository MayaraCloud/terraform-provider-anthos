package k8s

import (
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// InstallOrUpdateGKEConnectAgent installs or update a gke-connect agent in a Kubernetes cluster
// TODO: try to simplify the whole thing using restMapper and dynamic client
func InstallOrUpdateGKEConnectAgent(ctx context.Context, auth Auth, manifestResponse ConnectManifestResponse, GCPSAKey string, namespace string) error {
	kubeClient, err := KubeClientSet(auth)
	if err != nil {
		return fmt.Errorf("Initializing Kube clientset: %w", err)
	}

	for _, manifest := range manifestResponse.Manifest {
		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode([]byte(manifest.Manifest), nil, nil)

		if err != nil {
			// One of the manifests is an empty object, but it is marked as a Secret, we need to populate it
			// with the GCP SA key contents
			if strings.Contains(err.Error(), "is missing") && manifest.Type.Kind == "Secret" {
				o := CreateGCPCredsSecret(GCPSAKey, namespace)
				_, err := kubeClient.CoreV1().Secrets(namespace).Get(ctx, o.Name, metav1.GetOptions{})
				if err != nil {
					if errors.IsNotFound(err) {
						_, err = kubeClient.CoreV1().Secrets(namespace).Create(ctx, &o, metav1.CreateOptions{})
						if err != nil {
							return fmt.Errorf("Creating secret %v, error was %w", manifest.Manifest, err)
						}
					} else {
						return fmt.Errorf("Getting secret: %v, error was %w", manifest.Manifest, err)
					}
				}
				// There is no version metadata on the secret, so we just update it if it already exists
				_, err = kubeClient.CoreV1().Secrets(namespace).Update(ctx, &o, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Updating secret %v, error was %w", manifest.Manifest, err)
				}
			} else {
				return fmt.Errorf("Error while decoding YAML object %v, error was: %w", manifest.Manifest, err)
			}
		}

		// now use switch over the type of the object
		// and match each type-case
		switch o := obj.(type) {
		case *v1.Namespace:
			namespace, err := kubeClient.CoreV1().Namespaces().Get(ctx, o.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = kubeClient.CoreV1().Namespaces().Create(ctx, o, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Creating namespace %v, error was %w", manifest.Manifest, err)
					}
				} else {
					return fmt.Errorf("Getting namespace: %v, error was %w", manifest.Manifest, err)
				}
			}
			if namespace.ObjectMeta.Labels["version"] != o.ObjectMeta.Labels["version"] {
				_, err = kubeClient.CoreV1().Namespaces().Update(ctx, o, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Updating namespace %v, error was %w", manifest.Manifest, err)
				}
			}
		case *v1.ServiceAccount:
			sa, err := kubeClient.CoreV1().ServiceAccounts(o.Namespace).Get(ctx, o.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = kubeClient.CoreV1().ServiceAccounts(o.Namespace).Create(ctx, o, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Creating service account %v, error was %w", manifest.Manifest, err)
					}
				} else {
					return fmt.Errorf("Getting service account: %v, error was %w", manifest.Manifest, err)
				}
			}
			if sa.ObjectMeta.Labels["version"] != o.ObjectMeta.Labels["version"] {
				_, err = kubeClient.CoreV1().ServiceAccounts(o.Namespace).Update(ctx, o, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Updating service account %v, error was %w", manifest.Manifest, err)
				}
			}
		case *rbacv1.Role:
			role, err := kubeClient.RbacV1().Roles(o.Namespace).Get(ctx, o.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = kubeClient.RbacV1().Roles(o.Namespace).Create(ctx, o, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Creating role %v, error was %w", manifest.Manifest, err)
					}
				} else {
					return fmt.Errorf("Getting role: %v, error was %w", manifest.Manifest, err)
				}
			}
			if role.ObjectMeta.Labels["version"] != o.ObjectMeta.Labels["version"] {
				_, err = kubeClient.RbacV1().Roles(o.Namespace).Update(ctx, o, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Updating role %v, error was %w", manifest.Manifest, err)
				}
			}
		case *rbacv1.RoleBinding:
			roleBinding, err := kubeClient.RbacV1().RoleBindings(o.Namespace).Get(ctx, o.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = kubeClient.RbacV1().RoleBindings(o.Namespace).Create(ctx, o, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Creating role binding %v, error was %w", manifest.Manifest, err)
					}
				} else {
					return fmt.Errorf("Getting role binding: %v, error was %w", manifest.Manifest, err)
				}
			}
			if roleBinding.ObjectMeta.Labels["version"] != o.ObjectMeta.Labels["version"] {
				_, err = kubeClient.RbacV1().RoleBindings(o.Namespace).Update(ctx, o, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Updating role binding %v, error was %w", manifest.Manifest, err)
				}
			}
		case *rbacv1.ClusterRole:
			clusterRole, err := kubeClient.RbacV1().ClusterRoles().Get(ctx, o.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = kubeClient.RbacV1().ClusterRoles().Create(ctx, o, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Creating cluster role %v, error was %w", manifest.Manifest, err)
					}
				} else {
					return fmt.Errorf("Getting cluster role: %v, error was %w", manifest.Manifest, err)
				}
			}
			if clusterRole.ObjectMeta.Labels["version"] != o.ObjectMeta.Labels["version"] {
				_, err = kubeClient.RbacV1().ClusterRoles().Update(ctx, o, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Updating cluster role %v, error was %w", manifest.Manifest, err)
				}
			}
		case *rbacv1.ClusterRoleBinding:
			clusterRoleBinding, err := kubeClient.RbacV1().ClusterRoleBindings().Get(ctx, o.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = kubeClient.RbacV1().ClusterRoleBindings().Create(ctx, o, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Creating cluster role binding %v, error was %w", manifest.Manifest, err)
					}
				} else {
					return fmt.Errorf("Getting cluster role binding: %v, error was %w", manifest.Manifest, err)
				}
			}
			if clusterRoleBinding.ObjectMeta.Labels["version"] != o.ObjectMeta.Labels["version"] {
				_, err = kubeClient.RbacV1().ClusterRoleBindings().Update(ctx, o, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Updating cluster role binding %v, error was %w", manifest.Manifest, err)
				}
			}
		case *v1.Service:
			service, err := kubeClient.CoreV1().Services(o.Namespace).Get(ctx, o.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = kubeClient.CoreV1().Services(o.Namespace).Create(ctx, o, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Creating service %v, error was %w", manifest.Manifest, err)
					}
				} else {
					return fmt.Errorf("Getting service: %v, error was %w", manifest.Manifest, err)
				}
			}
			if service.ObjectMeta.Labels["version"] != o.ObjectMeta.Labels["version"] {
				_, err = kubeClient.CoreV1().Services(o.Namespace).Update(ctx, o, metav1.UpdateOptions{})
				if err != nil {
					if errors.IsInvalid(err) {
						err = kubeClient.CoreV1().Services(o.Namespace).Delete(ctx, o.Name, metav1.DeleteOptions{})
						if err != nil {
							return fmt.Errorf("Updating service %v, error was %w", manifest.Manifest, err)
						}
						_, err = kubeClient.CoreV1().Services(o.Namespace).Create(ctx, o, metav1.CreateOptions{})
						if err != nil {
							return fmt.Errorf("Creating service %v, error was %w", manifest.Manifest, err)
						}
					} else {
						return fmt.Errorf("Updating service %v, error was %w", manifest.Manifest, err)
					}
				}
			}
		case *appsv1.Deployment:
			deployment, err := kubeClient.AppsV1().Deployments(o.Namespace).Get(ctx, o.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = kubeClient.AppsV1().Deployments(o.Namespace).Create(ctx, o, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Creating deployment %v, error was %w", manifest.Manifest, err)
					}
				} else {
					return fmt.Errorf("Getting deployment: %v, error was %w", manifest.Manifest, err)
				}
			}
			if deployment.ObjectMeta.Labels["version"] != o.ObjectMeta.Labels["version"] {
				_, err = kubeClient.AppsV1().Deployments(o.Namespace).Update(ctx, o, metav1.UpdateOptions{})
				if err != nil {
					if errors.IsInvalid(err) {
						err = kubeClient.AppsV1().Deployments(o.Namespace).Delete(ctx, o.Name, metav1.DeleteOptions{})
						if err != nil {
							return fmt.Errorf("Updating deployment %v, error was %w", manifest.Manifest, err)
						}
						_, err = kubeClient.AppsV1().Deployments(o.Namespace).Create(ctx, o, metav1.CreateOptions{})
						if err != nil {
							return fmt.Errorf("Creating deployment %v, error was %w", manifest.Manifest, err)
						}
					} else {
						return fmt.Errorf("Updating deployment %v, error was %w", manifest.Manifest, err)
					}
				}
			}
		case *v1.Secret:
			_, err := kubeClient.CoreV1().Secrets(namespace).Get(ctx, o.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = kubeClient.CoreV1().Secrets(namespace).Create(ctx, o, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Creating secret %v, error was %w", manifest.Manifest, err)
					}
				} else {
					return fmt.Errorf("Getting secret: %v, error was %w", manifest.Manifest, err)
				}
			}
			// There is no version metadata on the secret, so we just update it if it already exists
			_, err = kubeClient.CoreV1().Secrets(namespace).Update(ctx, o, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("Updating secret %v, error was %w", manifest.Manifest, err)
			}
		}
	}

	return nil
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

// CreateGCPCredsSecret creates a kubernetes secret with a GCP Service Account key
func CreateGCPCredsSecret(GCPSAKey string, namespace string) v1.Secret {
	var secret v1.Secret
	secret.Data = make(map[string][]byte)
	secret.Name = "creds-gcp"
	secret.Namespace = namespace
	secret.Data["creds-gcp"] = []byte(GCPSAKey)
	return secret
}
