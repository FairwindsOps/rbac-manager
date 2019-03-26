# RBAC Definitions
Just as Kubernetes Deployments make Pods much simpler to manage at scale, RBAC Definitions are designed to simplify the management of Role Bindings and Service Accounts at scale. RBAC Manager will create, update, or delete Cluster Role Bindings, Role Bindings, or Service Accounts that are referenced in an RBAC Definition. Here's a more complete example of what that could look like:

```yaml
apiVersion: rbacmanager.reactiveops.io/v1beta1
kind: RBACDefinition
metadata:
  name: rbac-manager-users-example
rbacBindings:
  - name: cluster-admins
    subjects:
      - kind: User
        name: jane@example.com
    clusterRoleBindings:
      - clusterRole: cluster-admin
  - name: web-developers
    subjects:
      - kind: User
        name: dave@example.com
      - kind: User
        name: joe@example.com
    roleBindings:
      - clusterRole: edit
        namespace: web
      - clusterRole: view
        namespace: api
  - name: ci-bot
    subjects:
      - kind: ServiceAccount
        name: ci-bot
        namespace: rbac-manager
    roleBindings:
      - clusterRole: edit
        namespaceSelector:
          matchLabels:
            ci: edit
      - clusterRole: admin
        namespaceSelector:
          matchExpressions:
            - key: app
              operator: In
              values:
                - web
                - queue
```

In the above example, RBAC Manager will create the following resources:
- A Cluster Role Binding that gives Jane cluster-admin access
- A Role Binding that gives Dave and Joe edit access in the web namespace
- A Role Binding that gives Dave and Joe view access in the api namespace
- A Service Account named ci-bot in the rbac-manager namespace
- Role Binding(s) that grant the ci-bot Service Account edit access in all namespaces with `ci=edit` labels
- Role Binding(s) that grant the ci-bot Service Account admin access in all namespaces with `app=web` or `app=queue` labels

There are more examples of RBAC Definitions in the examples directory of this repo.
