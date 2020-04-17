resource "anthos_cluster_membership" "mayara_eks" {
    cluster_name = "gke_cluster"
    k8s_context = "mayara-eks"
}

resource "anthos_cluster_membership" "mayara_gke" {
    cluster_name = "gke_cluster"
    k8s_context = "mayara-gke"
}