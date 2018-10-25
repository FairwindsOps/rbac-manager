# Authentication With Azure Kubernetes Service
Azure Kubernetes Service (AKS) enables mapping Azure Active Directory (AAD) users and groups to RBAC users and groups. Although not enabled by default, this post shows [how to enable AAD mapping to RBAC](https://docs.microsoft.com/en-us/azure/aks/aad-integration).

## A Simple RBAC Definition for AKS
With AAD integration enabled, groups and users are mapped directly to RBAC. Users are mapped using their user name (generally email) while groups are mapped using the group object ID.

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
  - name: sample-group
    subjects:
      - kind: Group
        name: 894656e1-39f8-4bfe-b16a-510f61af6f41
    roleBindings:
      - clusterRole: edit
        namespace: api
```

Because group IDs are not easy to understand in a Kubernetes context, we generally recommend sticking with AAD users and using RBAC Definitions to group them. This ensures that all authorization config for Kubernetes stays within Kubernetes.