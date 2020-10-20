# Design and Architecture Notes

## cmd/manager/main.go

This is the primary entrypoint

## pkg/watcher

This package watches all resources that rbac-manager "owns" in order to trigger reconciliation if an outside actor modifies or deletes one of them

## pkg/reconciler/parser.go

Here the rbacDefinition is parsed into ServiceAccounts, ClusterRoleBindings, and RoleBindings

## pkg/controller

This package contains the watchers of Namesapces and RbacDefinitions, which are the primary things that can be used to trigger rbac-manager actions.

## pkg/reconciler/reconciler.go

This contains the functions that reconcile Namespaces, ServiceAccounts, ClusterRoleBindings, RoleBindings, and OnwerReferences

## pkg/apis

This contains the types necessary to define the RbacDefinition.
