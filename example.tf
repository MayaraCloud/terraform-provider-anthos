# This is just an example file

resource "anthos_cluster_membership" "mayara_eks" {
  cluster_name   = "mayara-eks"
  k8s_context    = "mayara-eks"
  description    = "mayara AWS based k8s cluster"
  hub_project_id = "mayara-anthos"
}

resource "anthos_gke_connect_agent" "mayara_eks_agent" {
  cluster_name = anthos_cluster_membership.mayara_eks.cluster_name
  k8s_context  = anthos_cluster_membership.mayara_eks.k8s_context
  description  = "mayara AWS based k8s gke connect agent"
  project      = anthos_cluster_membership.mayara_eks.hub_project_id
}
