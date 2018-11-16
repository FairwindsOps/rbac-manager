# Authentication With AWS
On AWS, we recommend using [aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator) for Kubernetes authentication. With EKS, this is included by default, and it's fairly straightforward to [setup with Kops](https://github.com/kubernetes-sigs/aws-iam-authenticator#kops-usage) or other methods of cluster provisioning. This library provides a variety of ways of mapping IAM roles and users to Kubernetes groups and users.

In the examples below, we'll show how aws-iam-authenticator configuration can work with rbac-manager. In all cases, the aws-iam-authenticator configuration snippets will represent part of the Kubernetes ConfigMap it reads config from. With EKS, the ConfigMap is named `aws-auth`, though other deployment patterns may use different naming.

```
kubectl get configmap -n kube-system aws-auth -oyaml
```

More [information about configuring aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator#full-configuration-format) is available in the official readme.

## Mapping Roles to Groups
One of the most common uses of aws-iam-authenticator involves mapping the AWS IAM Roles to Kubernetes Groups. The examples here use a bit of a shortcut to use a group that already is bound to a cluster-admin role (`system:masters`).

```yaml
mapRoles:
  - roleARN: arn:aws:iam::000000000000:role/KubernetesAdmin
    username: kubernetes-admin
    groups:
    - system:masters
```

Although this works, it's rather inelegant and doesn't allow you to modify the access this group has with RBAC without also affecting the system:masters group. A better alternative would be to map this to a new Kubernetes group that you can attach specific RBAC bindings to:

```yaml
mapRoles:
  - roleARN: arn:aws:iam::000000000000:role/KubernetesAdmin
    username: kubernetes-admin
    groups:
    - kubernetes-admins
```

With that group, you could create RBAC Bindings with RBAC Manager that would allow you to specify access specifically for that group.

```yaml
apiVersion: rbacmanager.reactiveops.io/v1beta1
kind: RBACDefinition
metadata:
  name: kubernetes-admins
rbacBindings:
  - name: kubernetes-admins
    subjects:
      - kind: Group
        name: kubernetes-admins
    clusterRoleBindings:
      - clusterRole: cluster-admin
```

A downside to this approach is that authorization configuration ends up getting split between AWS IAM Roles assigned to users and RBAC bindings. To understand what a user has access to in Kubernetes, you first have to determine what IAM roles they can assume, then what Kubernetes groups those roles map to, then what roles are bound to those Kubernetes groups. It can get quite complex to understand the level of access granted to your cluster.

## Mapping Specific Users
An alternative approach involves mapping specific IAM users to RBAC users. The configuration looks like this.

```yaml
mapUsers:
  - userARN: arn:aws:iam::012345678901:user/Jane
    username: jane
  - userARN: arn:aws:iam::012345678901:user/Joe
    username: joe
```

With the above config you could specify RBAC Bindings with the following RBAC Definition:

```yaml
apiVersion: rbacmanager.reactiveops.io/v1beta1
kind: RBACDefinition
metadata:
  name: sample-config
rbacBindings:
  - name: web-developers
    subjects:
      - kind: User
        name: jane
      - kind: User
        name: joe
    roleBindings:
      - clusterRole: edit
        namespace: web
```

Although we no longer need to understand which AWS IAM users can assume specific IAM roles. To understand what a user has access to in Kubernetes, you still need to understand what IAM users have been mapped to which usernames, then what roles are bound to those Kubernetes users. Although this is simpler, it still requires understanding 2 different sets of config.

## Mapping All Users in an AWS Account Automatically
If you like the above approach, but would prefer to just map all IAM users automatically to RBAC users, there's a configuration option for that as well:

```yaml
mapAccounts:
  - "012345678901"
```

With this approach, all IAM users are mapped to Kubernetes users with the full ARN as the username. By default, these users will be part of the `system:authenticated` group. That group is generally granted minimal permissions. With [rbac-lookup](http://github.com/reactiveops/rbac-lookup), you can view exactly what has been granted to that group with the following command:

```
rbac-lookup system:authenticated
```

To grant specific permissions to users we can use an RBAC Definition:

```yaml
apiVersion: rbacmanager.reactiveops.io/v1beta1
kind: RBACDefinition
metadata:
  name: sample-config
rbacBindings:
  - name: web-developers
    subjects:
      - kind: User
        name: arn:aws:iam::012345678901:user/Jane
      - kind: User
        name: arn:aws:iam::012345678901:user/Joe
    roleBindings:
      - clusterRole: edit
        namespace: web
```

By using `mapAccounts`, all authorization config lives within RBAC itself, allowing you to rely entirely on RBAC configuration for a clear picture of auth.