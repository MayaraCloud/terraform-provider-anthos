resource "anthos_cluster_membership" "mayara_eks" {
    cluster_name = "gke_cluster"
    k8s_context = "mayara-eks"
    description = "mayara AWS based k8s cluster"
    hub_project_id = "mayara-anthos"

}

#resource "anthos_cluster_membership" "mayara_gke" {
#    cluster_name = "gke_cluster"
#    k8s_context = "mayara-gke"
#    description = "mayara GKE based k8s cluster"
#    hub_project_id = "mayara-anthos"
#}