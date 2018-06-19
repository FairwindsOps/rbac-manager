# RBAC Manager

RBAC Manager simplifies the management of RBAC resources in Kubernetes. It has 3 primary goals:

1. Provide simplified RBAC configuration that will scale.
2. Use a syntax that can act as a centralized source of truth for RBAC configuration.
3. Enable automation of RBAC configuration changes.

## Introduction

Ideally RBAC Role Bindings should be configured to allow minimal access to a cluster. That generally means specifying access at a namespace and user level. For example, User A may need `edit` access to an api and web namespace. To create those role bindings, you'd need something like the following YAML configuration:

```
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: user-a-access
  namespace: web
subjects:
- kind: User
  name: A
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: edit
  apiGroup: rbac.authorization.k8s.io
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: user-a-access
  namespace: api
subjects:
- kind: User
  name: A
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: edit
  apiGroup: rbac.authorization.k8s.io
```

What's worse, to make User A an `admin` of Namespaces 1, we could not just update an existing Role Binding. Instead, that Role Binding would have to be deleted and replaced with a new one. With RBAC Manager, we can represent the state described above with some simpler YAML:

```
apiVersion: rbac-manager.reactiveops.io/v1beta1
kind: RbacDefinition
metadata:
  name: rbac-manager-config
spec:
  rbacBindings:
    - name: user-a
      subjects:
        - kind: User
          name: A
      roleBindings:
        - clusterRole: edit
          namespace: web
        - clusterRole: edit
          namespace: api
```

Of course, RBAC Manager is capable of so much more than that. It can manage Role Bindings and Cluster Role Bindings to Kubernetes Service Accounts, Users, and Groups. Here's a more complete example of a `RbacDefinition`:

```
apiVersion: rbac-manager.reactiveops.io/v1beta1
kind: RbacDefinition
metadata:
  name: rbac-manager-config
spec:
  rbacBindings:
    - name: example-group
      subjects:
        - kind: Group
          name: example
      clusterRoleBindings:
        - clusterRole: edit
      roleBindings:
        - clusterRole: admin
          namespace: default
    - name: example-users
      subjects:
        - kind: User
          name: sue@example.com
        - kind: User
          name: joe@example.com
      clusterRoleBindings:
        - clusterRole: edit
      roleBindings:
        - clusterRole: admin
          namespace: default
    - name: example-service-account
      subjects:
        - kind: ServiceAccount
          name: example
          namespace: default
      clusterRoleBindings:
        - clusterRole: view
      roleBindings:
        - clusterRole: admin
          namespace: default
```

RBAC Manager treats an `RbacDefinition` as a source of truth. All resources created by RBAC Manager are tied to the relevant `RbacDefinition` with an owner reference. Anytime Role Bindings are removed from a `RbacDefinition`, RBAC Manager will remove the associated Role Bindings that were created.

## Usage

RBAC Manager is a Kubernetes operator, powered by the [operator-sdk](https://github.com/operator-framework/operator-sdk). The simplest way to install this operator is with Helm, using the chart found in this repository.

```
helm install reactiveops/rbac-manager
```

Once RBAC Manager is installed in your cluster, you'll be able to deploy an `RbacDefinition` to configure your RBAC Bindings. Here's an example of a simple `RbacDefinition`:

```
apiVersion: rbac-manager.reactiveops.io/v1beta1
kind: RbacDefinition
metadata:
  name: rbac-manager-config
spec:
  rbacBindings:
    - name: api-developers
      subjects:
        - kind: User
          name: sue@example.com
        - kind: User
          name: joe@example.com
      clusterRoleBindings:
        - clusterRole: view
      roleBindings:
        - clusterRole: admin
          namespace: api
        - clusterRole: edit
          namespace: web
```

There are some additional sample `RbacDefinitions` in the `examples` directory.

### Deploying an RbacDefinition with the Helm chart

You can use the Helm chart to deploy the `RbacDefinition` along with the controller by adding these values:

```
rbacDefinition:
  enabled: true
  rbacBindings:
    - name: example-service-account
      subjects:
        - kind: ServiceAccount
          name: example
          namespace: default
      clusterRoleBindings:
        - clusterRole: view
      roleBindings:
        - clusterRole: admin
          namespace: default
```

## License
Apache License 2.0
