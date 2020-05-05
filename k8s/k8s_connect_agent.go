package k8s

import (
	"context"
	"fmt"

	"github.com/MayaraCloud/terraform-provider-anthos/debug"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// InstallOrUpdateGKEConnectAgent installs or update a gke-connect agent in a Kubernetes cluster
func InstallOrUpdateGKEConnectAgent(ctx context.Context, auth Auth, manifestResponse ConnectManifestResponse) error {
	kubeClient, err := KubeClientSet(auth)
	if err != nil {
		return fmt.Errorf("Initializing Kube clientset: %w", err)
	}

	for _, manifest := range manifestResponse.Manifest {
		decode := scheme.Codecs.UniversalDeserializer().Decode
		debug.GoLog(manifest.Type.Kind)
		obj, _, err := decode([]byte(manifest.Manifest), nil, nil)

		if err != nil {
			return fmt.Errorf("Error while decoding YAML object %v, error was: %w", manifest.Manifest, err)
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
						return fmt.Errorf("Creating namespace %w", err)
					}
				} else {
					return fmt.Errorf("Getting namespace: %w", err)
				}
			}
			if namespace.ObjectMeta.Labels["version"] != o.ObjectMeta.Labels["version"] {
				_, err = kubeClient.CoreV1().Namespaces().Update(ctx, o, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Updating namespace %w", err)
				}
			}
		case *v1.ServiceAccount:
			sa, err := kubeClient.CoreV1().ServiceAccounts(o.Namespace).Get(ctx, o.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = kubeClient.CoreV1().ServiceAccounts(o.Namespace).Create(ctx, o, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Creating service account %w", err)
					}
				} else {
					return fmt.Errorf("Getting service account: %w", err)
				}
			}
			if sa.ObjectMeta.Labels["version"] != o.ObjectMeta.Labels["version"] {
				_, err = kubeClient.CoreV1().ServiceAccounts(o.Namespace).Update(ctx, o, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Updating service account %w", err)
				}
			}
		case *rbacv1.Role:
			role, err := kubeClient.RbacV1().Roles(o.Namespace).Get(ctx, o.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = kubeClient.RbacV1().Roles(o.Namespace).Create(ctx, o, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Creating role %w", err)
					}
				} else {
					return fmt.Errorf("Getting role: %w", err)
				}
			}
			if role.ObjectMeta.Labels["version"] != o.ObjectMeta.Labels["version"] {
				_, err = kubeClient.RbacV1().Roles(o.Namespace).Update(ctx, o, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Updating role %w", err)
				}
			}
		case *rbacv1.RoleBinding:
			roleBinding, err := kubeClient.RbacV1().RoleBindings(o.Namespace).Get(ctx, o.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = kubeClient.RbacV1().RoleBindings(o.Namespace).Create(ctx, o, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Creating role binding %w", err)
					}
				} else {
					return fmt.Errorf("Getting role binding: %w", err)
				}
			}
			if roleBinding.ObjectMeta.Labels["version"] != o.ObjectMeta.Labels["version"] {
				_, err = kubeClient.RbacV1().RoleBindings(o.Namespace).Update(ctx, o, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Updating role binding %w", err)
				}
			}
		case *rbacv1.ClusterRole:
			clusterRole, err := kubeClient.RbacV1().ClusterRoles().Get(ctx, o.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = kubeClient.RbacV1().ClusterRoles().Create(ctx, o, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Creating cluster role %w", err)
					}
				} else {
					return fmt.Errorf("Getting cluster role: %w", err)
				}
			}
			if clusterRole.ObjectMeta.Labels["version"] != o.ObjectMeta.Labels["version"] {
				_, err = kubeClient.RbacV1().ClusterRoles().Update(ctx, o, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Updating cluster role %w", err)
				}
			}
		case *rbacv1.ClusterRoleBinding:
			clusterRoleBinding, err := kubeClient.RbacV1().ClusterRoleBindings().Get(ctx, o.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = kubeClient.RbacV1().ClusterRoleBindings().Create(ctx, o, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Creating cluster role binding %w", err)
					}
				} else {
					return fmt.Errorf("Getting cluster role binding: %w", err)
				}
			}
			if clusterRoleBinding.ObjectMeta.Labels["version"] != o.ObjectMeta.Labels["version"] {
				_, err = kubeClient.RbacV1().ClusterRoleBindings().Update(ctx, o, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Updating cluster role binding %w", err)
				}
			}
		case *v1.Service:
			service, err := kubeClient.CoreV1().Services(o.Namespace).Get(ctx, o.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = kubeClient.CoreV1().Services(o.Namespace).Create(ctx, o, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Creating service %w", err)
					}
				} else {
					return fmt.Errorf("Getting service: %w", err)
				}
			}
			if service.ObjectMeta.Labels["version"] != o.ObjectMeta.Labels["version"] {
				_, err = kubeClient.CoreV1().Services(o.Namespace).Update(ctx, o, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Updating service %w", err)
				}
			}
		case *appsv1.Deployment:
			deployment, err := kubeClient.AppsV1().Deployments(o.Namespace).Get(ctx, o.Name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					_, err = kubeClient.AppsV1().Deployments(o.Namespace).Create(ctx, o, metav1.CreateOptions{})
					if err != nil {
						return fmt.Errorf("Creating deployment %w", err)
					}
				} else {
					return fmt.Errorf("Getting deployment: %w", err)
				}
			}
			if deployment.ObjectMeta.Labels["version"] != o.ObjectMeta.Labels["version"] {
				_, err = kubeClient.AppsV1().Deployments(o.Namespace).Update(ctx, o, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Updating deployment %w", err)
				}
			}
			//case *v1.Secret:
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
