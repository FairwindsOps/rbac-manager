# Upgrades

## Upgrading to RBAC Manager 0.4.0
RBAC Manager 0.4.0 was a huge release that unfortunately included some significant breaking changes. To our knowledge, we're the only ones that have been using this project up to this point, but we certainly still do not take breaking changes lightly. We hope that the advantages it provides are worth the upgrade effort. This version involves a transition from Python to Go + Kubebuilder. There are a number of advantages included in this move:

- Improved testing capabilities with client-go
- Decreases Docker image size by 2/3
- Encourages usage of best practices through framework usage
- Strong typing should help avoid introducing stupid bugs
- Potential for generating docs with Kubebuilder

Beyond those benefits, here are the key new changes involved with this release:

- New RBAC Definition syntax
- Support for multiple RBAC Definitions
- Using OwnerReferences to associate a specific RBAC Definition with a Cluster Role Binding, Role Binding, or Service Account
- Using CRD validation for all RBAC Definition fields
- Moving from rbac-manager.reactiveops.io to rbacmanager.reactiveops.io for CRD
- Support for groups (helpful when working with something like Heptio Authenticator)
- Some basic testing in place

### The Upgrade Process
Unfortunately, with the changes to the custom resource definition, this process will essentially involve deleting the previous configuration and creating new replacement configuration with the updated syntax. Before beginning this process, ensure that you have cluster-admin access that is not managed by RBAC Manager. Part of this process will involve temporarily deleting those roles, so it's important to ensure you have cluster-admin access through a separate Cluster Role Binding.

1. Delete any existing RBAC Definitions - this allows the original RBAC Manager version to delete any associated role bindings.
2. Delete the RBAC Manager deployment configuration.
3. Deploy the updated version of RBAC Manager.
4. Create an updated RBAC Definition with the new syntax (see the `examples` directory of this repository).