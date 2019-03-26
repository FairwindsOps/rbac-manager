# Authentication With Google Kubernetes Engine
Google Kubernetes Engine (GKE) takes a unique approach to auth. Google IAM users are automatically mapped to Kubernetes RBAC users. Unfortunately there is no mapping for IAM groups to RBAC groups with GKE at this point.

## Initial RBAC Setup on GKE
The first time you configure RBAC on a GKE cluster, you may need to first grant yourself RBAC access. That can be accomplished with a command like this:

```
kubectl create clusterrolebinding initial-cluster-admin \
    --clusterrole=cluster-admin \
    --user=$(gcloud config get-value account)
```

## A Simple RBAC Definition for GKE
Google IAM users are mapped to Kubernetes RBAC users with their email as the username. This is also the case for Google IAM Service accounts. That makes RBAC Bindings very straightforward:

```yaml
apiVersion: rbacmanager.reactiveops.io/v1beta1
kind: RBACDefinition
metadata:
  name: sample-config
rbacBindings:
  - name: web-developers
    subjects:
      - kind: User
        name: jane@example.com
      - kind: User
        name: joe@example.com
    roleBindings:
      - clusterRole: edit
        namespace: web
```

## Understanding the Overlap Between IAM and RBAC
Google Cloud IAM roles can provide fairly specific Kubernetes authorization configuration that overlaps with RBAC roles. This means that a user's access to a GKE cluster ends up being a union of both IAM and RBAC roles. This blog post provides more information on [how IAM and RBAC work together in GKE](https://medium.com/uptime-99/making-sense-of-kubernetes-rbac-and-iam-roles-on-gke-914131b01922). If you're simply trying to see relevant GKE IAM and RBAC roles in one place, [rbac-lookup can help](https://github.com/reactiveops/rbac-lookup) with that.
