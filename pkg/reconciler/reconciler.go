// Copyright 2018 FairwindsOps Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reconciler

import (
	"context"
	"log/slog"
	"reflect"
	"sync"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"

	rbacmanagerv1beta1 "github.com/fairwindsops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
	"github.com/fairwindsops/rbac-manager/pkg/kube"
	"github.com/fairwindsops/rbac-manager/pkg/metrics"
)

// Reconciler creates and deletes Kubernetes resources to achieve the desired state of an RBAC Definition
type Reconciler struct {
	Clientset kubernetes.Interface
	ownerRefs []metav1.OwnerReference
}

var mux = sync.Mutex{}

// ReconcileNamespaceChange reconciles relevant portions of RBAC Definitions
//
//	after changes to namespaces within the cluster
func (r *Reconciler) ReconcileNamespaceChange(rbacDef *rbacmanagerv1beta1.RBACDefinition, namespace *v1.Namespace) error {
	mux.Lock()
	defer mux.Unlock()

	r.ownerRefs = rbacDefOwnerRefs(rbacDef)

	p := Parser{
		Clientset: r.Clientset,
		ownerRefs: r.ownerRefs,
	}

	err := p.Parse(*rbacDef)
	if err != nil {
		return err
	}

	err = r.reconcileServiceAccounts(&p.parsedServiceAccounts)
	if err != nil {
		return err
	}

	if p.hasNamespaceSelectors(rbacDef) {
		slog.Info("Reconciling namespace", "namespace", namespace.Name, "rbacDefinition", rbacDef.Name)
		err := r.reconcileRoleBindings(&p.parsedRoleBindings)
		if err != nil {
			return err
		}
	}

	return nil
}

// ReconcileOwners reconciles any RBACDefinitions found in owner references
func (r *Reconciler) ReconcileOwners(ownerRefs []metav1.OwnerReference, kind string) error {
	mux.Lock()
	defer mux.Unlock()

	namespaces, err := r.Clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		slog.Debug("Error listing namespaces", "error", err)
		return err
	}

	for _, ownerRef := range ownerRefs {
		if ownerRef.Kind == "RBACDefinition" {
			rbacDef, err := kube.GetRbacDefinition(ownerRef.Name)
			if err != nil {
				return err
			}

			r.ownerRefs = rbacDefOwnerRefs(&rbacDef)

			p := Parser{
				Clientset: r.Clientset,
				ownerRefs: r.ownerRefs,
			}

			if kind == "RoleBinding" {
				p.parseRoleBindings(&rbacDef, namespaces)
				return r.reconcileRoleBindings(&p.parsedRoleBindings)
			} else if kind == "ClusterRoleBinding" {
				p.parseClusterRoleBindings(&rbacDef)
				return r.reconcileClusterRoleBindings(&p.parsedClusterRoleBindings)
			} else if kind == "ServiceAccount" {
				err := p.Parse(rbacDef)
				if err != nil {
					return err
				}
				return r.reconcileServiceAccounts(&p.parsedServiceAccounts)
			}
		}
	}
	return nil
}

// Reconcile creates, updates, or deletes Kubernetes resources to match
//
//	the desired state defined in an RBAC Definition
func (r *Reconciler) Reconcile(rbacDef *rbacmanagerv1beta1.RBACDefinition) error {
	mux.Lock()
	defer mux.Unlock()

	slog.Info("Reconciling RBACDefinition", "name", rbacDef.Name)

	r.ownerRefs = rbacDefOwnerRefs(rbacDef)

	p := Parser{
		Clientset: r.Clientset,
		ownerRefs: r.ownerRefs,
	}

	var err error

	err = p.Parse(*rbacDef)
	if err != nil {
		return err
	}

	err = r.reconcileServiceAccounts(&p.parsedServiceAccounts)
	if err != nil {
		return err
	}

	err = r.reconcileClusterRoleBindings(&p.parsedClusterRoleBindings)
	if err != nil {
		return err
	}

	err = r.reconcileRoleBindings(&p.parsedRoleBindings)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) reconcileServiceAccounts(requested *[]v1.ServiceAccount) error {
	existing, err := r.Clientset.CoreV1().ServiceAccounts("").List(context.TODO(), kube.ListOptions)
	if err != nil {
		return err
	}

	matchingServiceAccounts := []v1.ServiceAccount{}
	serviceAccountsToCreate := []v1.ServiceAccount{}

	for _, requestedSA := range *requested {
		alreadyExists := false
		for _, existingSA := range existing.Items {
			if saMatches(&existingSA, &requestedSA) {
				alreadyExists = true
				matchingServiceAccounts = append(matchingServiceAccounts, existingSA)
				break
			}
		}

		if !alreadyExists {
			serviceAccountsToCreate = append(serviceAccountsToCreate, requestedSA)
		} else {
			slog.Debug("Service Account already exists", "name", requestedSA.Name)
		}
	}

	for _, existingSA := range existing.Items {
		if reflect.DeepEqual(existingSA.ObjectMeta.OwnerReferences, r.ownerRefs) {
			matchingRequest := false
			for _, matchingSA := range matchingServiceAccounts {
				if saMatches(&existingSA, &matchingSA) {
					matchingRequest = true
					break
				}
			}

			if !matchingRequest {
				slog.Info("Deleting Service Account", "name", existingSA.Name)
				err := r.Clientset.CoreV1().ServiceAccounts(existingSA.Namespace).Delete(context.TODO(), existingSA.Name, metav1.DeleteOptions{})
				if err != nil {
					slog.Info("Error deleting Service Account", "name", existingSA.Name, "error", err)
					metrics.ErrorCounter.Inc()
				} else {
					metrics.ChangeCounter.WithLabelValues("serviceaccounts", "delete").Inc()
				}
			} else {
				slog.Debug("Matches requested Service Account", "name", existingSA.Name)
			}
		}
	}

	for _, serviceAccountToCreate := range serviceAccountsToCreate {
		slog.Info("Creating Service Account", "name", serviceAccountToCreate.Name)
		_, err := r.Clientset.CoreV1().ServiceAccounts(serviceAccountToCreate.ObjectMeta.Namespace).Create(context.TODO(), &serviceAccountToCreate, metav1.CreateOptions{})
		if err != nil {
			slog.Error("Error creating Service Account", "name", serviceAccountToCreate.Name, "error", err)
			metrics.ErrorCounter.Inc()
		} else {
			metrics.ChangeCounter.WithLabelValues("serviceaccounts", "create").Inc()
		}
	}

	return nil
}

func (r *Reconciler) reconcileClusterRoleBindings(requested *[]rbacv1.ClusterRoleBinding) error {
	existing, err := r.Clientset.RbacV1().ClusterRoleBindings().List(context.TODO(), kube.ListOptions)
	if err != nil {
		metrics.ErrorCounter.Inc()
		return err
	}

	matchingClusterRoleBindings := []rbacv1.ClusterRoleBinding{}
	clusterRoleBindingsToCreate := []rbacv1.ClusterRoleBinding{}

	for _, requestedCRB := range *requested {
		alreadyExists := false
		for _, existingCRB := range existing.Items {
			if crbMatches(&existingCRB, &requestedCRB) {
				alreadyExists = true
				matchingClusterRoleBindings = append(matchingClusterRoleBindings, existingCRB)
				break
			}
		}

		if !alreadyExists {
			clusterRoleBindingsToCreate = append(clusterRoleBindingsToCreate, requestedCRB)
		} else {
			slog.Debug("Cluster Role Binding already exists", "name", requestedCRB.Name)
		}
	}

	for _, existingCRB := range existing.Items {
		if reflect.DeepEqual(existingCRB.OwnerReferences, r.ownerRefs) {
			matchingRequest := false
			for _, requestedCRB := range matchingClusterRoleBindings {
				if crbMatches(&existingCRB, &requestedCRB) {
					matchingRequest = true
					break
				}
			}

			if !matchingRequest {
				slog.Info("Deleting Cluster Role Binding", "name", existingCRB.Name)
				err := r.Clientset.RbacV1().ClusterRoleBindings().Delete(context.TODO(), existingCRB.Name, metav1.DeleteOptions{})
				if err != nil {
					slog.Error("Error deleting Cluster Role Binding", "name", existingCRB.Name, "error", err)
					metrics.ErrorCounter.Inc()
				} else {
					metrics.ChangeCounter.WithLabelValues("clusterrolebindings", "delete").Inc()
				}
			} else {
				slog.Debug("Matches requested Cluster Role Binding", "name", existingCRB.Name)
			}
		}
	}

	for _, clusterRoleBindingToCreate := range clusterRoleBindingsToCreate {
		slog.Info("Creating Cluster Role Binding", "name", clusterRoleBindingToCreate.Name)
		_, err := r.Clientset.RbacV1().ClusterRoleBindings().Create(context.TODO(), &clusterRoleBindingToCreate, metav1.CreateOptions{})
		if err != nil {
			slog.Error("Error creating Cluster Role Binding", "name", clusterRoleBindingToCreate.Name, "error", err)
			metrics.ErrorCounter.Inc()
		} else {
			metrics.ChangeCounter.WithLabelValues("clusterrolebindings", "create").Inc()
		}
	}

	return nil
}

func (r *Reconciler) reconcileRoleBindings(requested *[]rbacv1.RoleBinding) error {
	existing, err := r.Clientset.RbacV1().RoleBindings("").List(context.TODO(), kube.ListOptions)
	if err != nil {
		return err
	}

	matchingRoleBindings := []rbacv1.RoleBinding{}
	roleBindingsToCreate := []rbacv1.RoleBinding{}

	for _, requestedRB := range *requested {
		alreadyExists := false
		for _, existingRB := range existing.Items {
			if rbMatches(&existingRB, &requestedRB) {
				alreadyExists = true
				matchingRoleBindings = append(matchingRoleBindings, existingRB)
				break
			}
		}

		if !alreadyExists {
			roleBindingsToCreate = append(roleBindingsToCreate, requestedRB)
		} else {
			slog.Debug("Role Binding already exists", "name", requestedRB.Name)
		}
	}

	for _, existingRB := range existing.Items {
		if reflect.DeepEqual(existingRB.OwnerReferences, r.ownerRefs) {
			matchingRequest := false
			for _, requestedRB := range matchingRoleBindings {
				if rbMatches(&existingRB, &requestedRB) {
					matchingRequest = true
					break
				}
			}

			if !matchingRequest {
				slog.Info("Deleting Role Binding", "name", existingRB.Name)
				err := r.Clientset.RbacV1().RoleBindings(existingRB.Namespace).Delete(context.TODO(), existingRB.Name, metav1.DeleteOptions{})
				if err != nil {
					slog.Info("Error deleting Role Binding", "name", existingRB.Name, "error", err)
					metrics.ErrorCounter.Inc()
				} else {
					metrics.ChangeCounter.WithLabelValues("rolebindings", "delete").Inc()
				}
			} else {
				slog.Debug("Matches requested Role Binding", "name", existingRB.Name)
			}
		}
	}

	for _, roleBindingToCreate := range roleBindingsToCreate {
		slog.Info("Creating Role Binding", "name", roleBindingToCreate.Name)
		_, err := r.Clientset.RbacV1().RoleBindings(roleBindingToCreate.ObjectMeta.Namespace).Create(context.TODO(), &roleBindingToCreate, metav1.CreateOptions{})
		if err != nil {
			slog.Error("Error creating Role Binding", "name", roleBindingToCreate.Name, "error", err)
			metrics.ErrorCounter.Inc()
		} else {
			metrics.ChangeCounter.WithLabelValues("rolebindings", "create").Inc()
		}
	}

	return nil
}

func rbacDefOwnerRefs(rbacDef *rbacmanagerv1beta1.RBACDefinition) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(rbacDef, schema.GroupVersionKind{
			Group:   rbacmanagerv1beta1.SchemeGroupVersion.Group,
			Version: rbacmanagerv1beta1.SchemeGroupVersion.Version,
			Kind:    "RBACDefinition",
		}),
	}
}
