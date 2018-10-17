# RBAC Manager

[![Go Report Card](https://goreportcard.com/badge/github.com/reactiveops/rbac-manager)](https://goreportcard.com/report/github.com/reactiveops/rbac-manager) [![CircleCI](https://circleci.com/gh/reactiveops/rbac-manager.svg?style=svg)](https://circleci.com/gh/reactiveops/rbac-manager)

RBAC Manager simplifies the management of Cluster Role Bindings, Role Bindings, and Service Accounts in Kubernetes. It has 3 primary goals:

1. Simplify RBAC in a secure and scalable approach. When it's easier to use, it will be used more often resulting in better security.
2. Declarative syntax that can be checked into source control and act as the source of truth for RBAC configuration.
3. Enable automation of RBAC configuration changes.

## Introduction

Ideally RBAC Role Bindings should be configured with the principle of least privilege in mind - users should have the minimum level of access possible. That generally means specifying access at a namespace and user level. For example, User A may need `edit` access to an api and web namespace. To create those role bindings, you'd need something like the following YAML configuration:

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

To make user A an `admin` of namespace 1, we are unable to update the existing Role Binding in place. Instead, we must delete the old binding and create a new binding to replace it. With RBAC Manager, we can represent the state described above with some simpler YAML:

```
apiVersion: rbacmanager.reactiveops.io/v1beta1
kind: RBACDefinition
metadata:
  name: rbac-manager-config
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

Of course, RBAC Manager is capable of so much more than that. It can manage Role Bindings and Cluster Role Bindings to Kubernetes Service Accounts, Users, and Groups. Here's a more complete example of a `RBACDefinition`:

```
apiVersion: rbacmanager.reactiveops.io/v1beta1
kind: RBACDefinition
metadata:
  name: rbac-manager-users-example
rbacBindings:
  - name: cluster-admins
    subjects:
      - kind: User
        name: sue@example.com
      - kind: User
        name: joe@example.com
    clusterRoleBindings:
      - clusterRole: cluster-admin
  - name: web-developers
    subjects:
      - kind: User
        name: sarah@example.com
      - kind: User
        name: john@example.com
      - kind: User
        name: daniel@example.com
    roleBindings:
      - clusterRole: edit
        namespace: web
      - clusterRole: view
        namespace: api
  - name: api-developers
    subjects:
      - kind: User
        name: jess@example.com
      - kind: User
        name: lance@example.com
      - kind: User
        name: rob@example.com
    roleBindings:
      - clusterRole: edit
        namespace: api
      - clusterRole: view
        namespace: web
  - name: ci-bot
    subjects:
      - kind: ServiceAccount
        name: ci-bot
    roleBindings:
      - clusterRole: edit
        namespace: api
      - clusterRole: edit
        namespace: web
```

RBAC Manager treats an `RBACDefinition` as the source of truth. All resources created by RBAC Manager are tied to the relevant `RBACDefinition` with an owner reference. If a desired role is changed in an RBACDefinition, the relevant Role Bindings are replaced with new bindings to requested role. Any time Role Bindings are removed from a `RBACDefinition`, RBAC Manager will also remove the associated Role Bindings that it had created. It's also worth noting that when a `ServiceAccount` is a subject, RBAC Manager will attempt to create the `ServiceAccount` if it doesn't already exist.

## Usage

RBAC Manager is a Kubernetes operator, powered by [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder). The simplest way to install this operator is with Helm, using the chart found in this repository.

```
helm install chart/ --name rbac-manager --namespace rbac-manager
```

Alternatively, the YAML template Helm generates are available in the `deploy` directory of this repo. If you'd prefer to deploy this directly with `kubectl`, you can do that with this command:

```
kubectl apply -f deploy/
```

Once RBAC Manager is installed in your cluster, you'll be able to deploy an `RBACDefinition` to configure your RBAC Bindings. Here's an example of a simple `RBACDefinition`:

```
apiVersion: rbacmanager.reactiveops.io/v1beta1
kind: RBACDefinition
metadata:
  name: rbac-manager-config
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

There are some additional sample `RBACDefinitions` in the `examples` directory.

### Deploying an RBACDefinition with the Helm chart

You can use the Helm chart to deploy the `RBACDefinition` along with the controller by adding these values:

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

## Note on upgrades to 0.4.0

This release has breaking changes. Please see our [upgrade process](docs/upgrades.md).

## RBAC Lookup

We have a related open source tool that allows you to easily find roles and cluster roles attached to any user, service account, or group name in your Kubernetes cluster. If that sounds interesting, take a look at [rbac-lookup](https://github.com/reactiveops/rbac-lookup) on GitHub.

## License
Apache License 2.0
