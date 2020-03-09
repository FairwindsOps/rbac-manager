<div align="center">
  <img src="/docs/logo.svg" height="200" alt="RBAC Manager" style="padding-bottom: 20px" />
  <br><br>

  [![Go Report Card](https://goreportcard.com/badge/github.com/FairwindsOps/rbac-manager)](https://goreportcard.com/report/github.com/FairwindsOps/rbac-manager) [![CircleCI](https://circleci.com/gh/FairwindsOps/rbac-manager.svg?style=svg)](https://circleci.com/gh/FairwindsOps/rbac-manager) [![codecov](https://codecov.io/gh/fairwindsops/rbac-manager/branch/master/graph/badge.svg)](https://codecov.io/gh/fairwindsops/rbac-manager)
</div>


RBAC Manager was designed to simplify authorization in Kubernetes. This is an operator that supports declarative configuration for RBAC with new custom resources. Instead of managing role bindings or service accounts directly, you can specify a desired state and RBAC Manager will make the necessary changes to achieve that state.

This project has three main goals:

1. Provide a declarative approach to RBAC that is more approachable and scalable.
2. Reduce the amount of configuration required for great auth.
3. Enable automation of RBAC configuration updates with CI/CD.

**Want to learn more?** Reach out on [the Slack channel](https://fairwindscommunity.slack.com/messages/rbac-projects), send an email to `opensource@fairwinds.com`, or join us for [office hours on Zoom](https://fairwindscommunity.slack.com/messages/office-hours)

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
  name: view
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

## The Benefits
With an RBAC Definition custom resource, we can cut the amount of configuration in half (or often significantly more). RBAC Manager is deployed as an operator and listens for new and updated RBAC Definitions, making the necessary changes to achieve the desired state.

This approach is incredibly helpful for 2 specific cases:

#### 1. Updating a Role Binding
Unfortunately there's no way to change the role an existing Kubernetes Role Binding refers to. That means that changing a role granted to a user involves deleting and recreating a Kubernetes Role Binding. With RBAC Manager, that process happens automatically when an RBAC Definition is updated.

#### 2. Detecting Role Binding Removal
When it comes to potential CI automation of changes to RBAC configuration, tracking the removal of a role binding can get quite tricky. If you were using a traditional workflow where desired Kubernetes objects are represent in a repo as yaml files, the creates and updates are reasonably straightforward, but revoking access on the basis of a Role Binding being removed is quite tricky.

With RBAC Manager, each RBAC Definition "owns" any resources it creates, and will always compare the desired state in the current RBAC Definition with the list of resources currently owned by it. If a Role Binding is no longer included in a RBAC Definition, RBAC Manager will automatically remove it.

## Getting Started
RBAC Manager is simple to install with either the [Helm chart](https://github.com/FairwindsOps/charts/tree/master/stable/rbac-manager) or Kubernetes deployment YAML included in this repo:

```
helm repo add fairwinds-stable https://charts.fairwinds.com/stable
helm install fairwinds-stable/rbac-manager --name rbac-manager --namespace rbac-manager
```

```
kubectl apply -f deploy/
```

Once RBAC Manager is installed in your cluster, you'll be able to deploy RBAC Definitions to your cluster. There are examples of these custom resources above as well as in the examples directory of this repository.

## Dynamic Namespaces and Labels
RBAC Definitions can now include `namespaceSelectors` in place of `namespace` attributes when specifying Role Binding configuration. This can be incredibly helpful when working with dynamically provisioned namespaces.

```yaml
apiVersion: rbacmanager.reactiveops.io/v1beta1
kind: RBACDefinition
metadata:
  name: dev-access
rbacBindings:
  - name: dev-team
    subjects:
      - kind: Group
        name: dev-team
    roleBindings:
      - clusterRole: edit
        namespaceSelector:
          matchLabels:
            team: dev
```

In the example above, Role Bindings would automatically get created for each Namespace with a `team=dev` label. This supports the same functionality as other Kubernetes label selectors, read the [official docs](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) for more information.

## ServiceAccounts

If an `RBACDefinition` defines a `ServiceAccount` as a subject, rbac-manager will attempt to create the `ServiceAccount` for you. **WARNING**: When an `RBACDefinition` owns a `ServiceAccount` in this fashion, it will be deleted when the `RBACDefinition` is deleted.

### ImagePullSecrets and ServiceAccounts

Service accounts support adding `ImagePullSecrets` to their definition. What happens is that when a `Pod` (via `Deployment` or otherwise) is launched specifying a `ServiceAccount` that specifies `ImagePullSecrets`, the pull secrets will be injected into the Pod spec automatically. An example of this using rbac-manager [can be found in the examples directory](examples/rbacdefinition-sa-imagepull.yaml).

Please note: rbac-manager will not manage secrets, and assumes they are already present in the same namespace that the `ServiceAccount` is in. Also, `ImagePullSecrets` only apply when the `Subject` is a `ServiceAccount`.

## Contributing
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Roadmap](ROADMAP.md)
- [Changelog](https://github.com/FairwindsOps/rbac-manager/releases)

## Further Reading

### RBAC Definitions
RBAC Definitions can manage Cluster Role Bindings, Role Bindings, and Service Accounts. To better understand how these work, read our [RBAC Definition documentation](docs/rbacdefinitions.md).

### Cloud Specific Authentication Tips
To properly configure authorization with RBAC in Kubernetes, you first need to have good authentication. We've provided some helpful documentation for working with authentication on [AWS](docs/aws.md), [Google Cloud](docs/gke.md), and [Azure](docs/aks.md).

### Better Visibility With RBAC Lookup
We have a related open source tool that allows you to easily find roles and cluster roles attached to any user, service account, or group name in your Kubernetes cluster. If that sounds interesting, take a look at [rbac-lookup](https://github.com/FairwindsOps/rbac-lookup) on GitHub.

## License
Apache License 2.0
