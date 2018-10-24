# Introduction
RBAC Manager was designed to simplify authorization in Kubernetes. This is an operator that supports declarative configuration for RBAC with new custom resources. Instead of managing role bindings or service accounts directly, you can specify a desired state and RBAC Manager will make the necessary changes to achieve that state.

This project has three main goals:

1. Provide a declarative approach to RBAC that is more approachable and scalable.
2. Reduce the amount of configuration required for great auth.
3. Enable automation of RBAC configuration updates with CI/CD.

## An Example
To fully understand how RBAC Manager works, it's helpful to walk through a simple example. In this example we'll have a single user, Joe, that needs `edit` access to the `web` namespace and `view` access to `api` namespace.

With RBAC, that requires creating 2 role bindings, the first grants `edit` access to the `web` namespace.
```yaml
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: joe-web
  namespace: web
subjects:
- kind: User
  name: joe@example.com
roleRef:
  kind: ClusterRole
  name: edit
  apiGroup: rbac.authorization.k8s.io
```

The second grants `view` access to the `api` namespace.
```yaml
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: joe-api
  namespace: api
subjects:
- kind: User
  name: joe@example.com
roleRef:
  kind: ClusterRole
  name: api
  apiGroup: rbac.authorization.k8s.io
```

It's easy to see just how repetitive this becomes. With RBAC Manager, we can use a custom resource to achieve the same result.
```yaml
apiVersion: rbacmanager.reactiveops.io/v1beta1
kind: RBACDefinition
metadata:
  name: joe-access
rbacBindings:
  - name: joe
    subjects:
      - kind: User
        name: joe@example.com
    roleBindings:
      - namespace: api
        clusterRole: view
      - namespace: web
        clusterRole: edit
```