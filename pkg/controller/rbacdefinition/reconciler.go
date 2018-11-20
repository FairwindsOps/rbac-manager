// Copyright 2018 ReactiveOps
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

package rbacdefinition

import (
	"reflect"

	rbacmanagerv1beta1 "github.com/reactiveops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
	logrus "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (bc *RBACDefinitionController) reconcileRbacDef(
	rbacDef *rbacmanagerv1beta1.RBACDefinition) error {

	logrus.Infof("Processing RBACDefinition %v", rbacDef.Name)

	ownerRefs := []metav1.OwnerReference{
		*metav1.NewControllerRef(rbacDef, schema.GroupVersionKind{
			Group:   rbacmanagerv1beta1.SchemeGroupVersion.Group,
			Version: rbacmanagerv1beta1.SchemeGroupVersion.Version,
			Kind:    "RBACDefinition",
		}),
	}

	rdp := rbacDefinitionParser{
		k8sClientSet: bc.kubernetesClientSet,
		labels:       map[string]string{"rbac-manager": "reactiveops"},
		listOptions:  metav1.ListOptions{LabelSelector: "rbac-manager=reactiveops"},
		ownerRefs:    ownerRefs,
	}

	err := rdp.parse(*rbacDef)
	if err != nil {
		return err
	}

	existingManagedClusterRoleBindings, err := bc.kubernetesClientSet.RbacV1().ClusterRoleBindings().List(rdp.listOptions)
	if err != nil {
		return err
	}

	existingManagedRoleBindings, err := bc.kubernetesClientSet.RbacV1().RoleBindings("").List(rdp.listOptions)
	if err != nil {
		return err
	}

	existingManagedServiceAccounts, err := bc.kubernetesClientSet.CoreV1().ServiceAccounts("").List(rdp.listOptions)
	if err != nil {
		return err
	}

	bc.reconcileServiceAccounts(
		&rdp.parsedServiceAccounts,
		&existingManagedServiceAccounts.Items,
		&rdp.ownerRefs)

	bc.reconcileClusterRoleBindings(
		&rdp.parsedClusterRoleBindings,
		&existingManagedClusterRoleBindings.Items,
		&rdp.ownerRefs)

	bc.reconcileRoleBindings(
		&rdp.parsedRoleBindings,
		&existingManagedRoleBindings.Items,
		&rdp.ownerRefs)

	return nil
}

func (bc *RBACDefinitionController) reconcileClusterRoleBindings(
	requestedClusterRoleBindings *[]rbacv1.ClusterRoleBinding,
	existingManagedClusterRoleBindings *[]rbacv1.ClusterRoleBinding,
	ownerReferences *[]metav1.OwnerReference) {

	matchingClusterRoleBindings := []rbacv1.ClusterRoleBinding{}
	clusterRoleBindingsToCreate := []rbacv1.ClusterRoleBinding{}

	for _, requestedCRB := range *requestedClusterRoleBindings {
		alreadyExists := false
		for _, existingCRB := range *existingManagedClusterRoleBindings {
			if crbMatches(&existingCRB, &requestedCRB) {
				alreadyExists = true
				matchingClusterRoleBindings = append(matchingClusterRoleBindings, existingCRB)
				break
			}
		}

		if !alreadyExists {
			clusterRoleBindingsToCreate = append(clusterRoleBindingsToCreate, requestedCRB)
		} else {
			logrus.Debugf("Cluster Role Binding already exists %v", requestedCRB.Name)
		}
	}

	for _, existingCRB := range *existingManagedClusterRoleBindings {
		if reflect.DeepEqual(existingCRB.OwnerReferences, *ownerReferences) {
			matchingRequest := false
			for _, requestedCRB := range matchingClusterRoleBindings {
				if crbMatches(&existingCRB, &requestedCRB) {
					matchingRequest = true
					break
				}
			}

			if !matchingRequest {
				logrus.Infof("Deleting Cluster Role Binding: %v", existingCRB.Name)
				err := bc.kubernetesClientSet.RbacV1().ClusterRoleBindings().Delete(existingCRB.Name, &metav1.DeleteOptions{})
				if err != nil {
					logrus.Errorf("Error deleting Cluster Role Binding: %v", err)
				}
			} else {
				logrus.Debugf("Matches requested Cluster Role Binding: %v", existingCRB.Name)
			}
		}
	}

	for _, clusterRoleBindingToCreate := range clusterRoleBindingsToCreate {
		logrus.Infof("Creating Cluster Role Binding: %v", clusterRoleBindingToCreate.Name)
		_, err := bc.kubernetesClientSet.RbacV1().ClusterRoleBindings().Create(&clusterRoleBindingToCreate)
		if err != nil {
			logrus.Errorf("Error creating Cluster Role Binding: %v", err)
		}
	}
}

func (bc *RBACDefinitionController) reconcileRoleBindings(
	requestedRoleBindings *[]rbacv1.RoleBinding,
	existingManagedRoleBindings *[]rbacv1.RoleBinding,
	ownerReferences *[]metav1.OwnerReference) {

	matchingRoleBindings := []rbacv1.RoleBinding{}
	roleBindingsToCreate := []rbacv1.RoleBinding{}

	for _, requestedRB := range *requestedRoleBindings {
		alreadyExists := false
		for _, existingRB := range *existingManagedRoleBindings {
			if rbMatches(&existingRB, &requestedRB) {
				alreadyExists = true
				matchingRoleBindings = append(matchingRoleBindings, existingRB)
				break
			}
		}

		if !alreadyExists {
			roleBindingsToCreate = append(roleBindingsToCreate, requestedRB)
		} else {
			logrus.Debugf("Role Binding already exists %v", requestedRB.Name)
		}
	}

	for _, existingRB := range *existingManagedRoleBindings {
		if reflect.DeepEqual(existingRB.OwnerReferences, *ownerReferences) {
			matchingRequest := false
			for _, requestedRB := range matchingRoleBindings {
				if rbMatches(&existingRB, &requestedRB) {
					matchingRequest = true
					break
				}
			}

			if !matchingRequest {
				logrus.Infof("Deleting Role Binding %v", existingRB.Name)
				err := bc.kubernetesClientSet.RbacV1().RoleBindings(existingRB.Namespace).Delete(existingRB.Name, &metav1.DeleteOptions{})
				if err != nil {
					logrus.Infof("Error deleting Role Binding: %v", err)
				}
			} else {
				logrus.Debugf("Matches requested Role Binding %v", existingRB.Name)
			}
		}
	}

	for _, roleBindingToCreate := range roleBindingsToCreate {
		logrus.Infof("Creating Role Binding: %v", roleBindingToCreate.Name)
		_, err := bc.kubernetesClientSet.RbacV1().RoleBindings(roleBindingToCreate.ObjectMeta.Namespace).Create(&roleBindingToCreate)
		if err != nil {
			logrus.Errorf("Error creating Role Binding: %v", err)
		}
	}
}

func (bc *RBACDefinitionController) reconcileServiceAccounts(
	requestedServiceAccounts *[]v1.ServiceAccount,
	existingManagedServiceAccounts *[]v1.ServiceAccount,
	ownerReferences *[]metav1.OwnerReference) {

	matchingServiceAccounts := []v1.ServiceAccount{}
	serviceAccountsToCreate := []v1.ServiceAccount{}

	for _, requestedSA := range *requestedServiceAccounts {
		alreadyExists := false
		for _, existingSA := range *existingManagedServiceAccounts {
			if saMatches(&existingSA, &requestedSA) {
				alreadyExists = true
				matchingServiceAccounts = append(matchingServiceAccounts, existingSA)
				break
			}
		}

		if !alreadyExists {
			serviceAccountsToCreate = append(serviceAccountsToCreate, requestedSA)
		} else {
			logrus.Debugf("Service Account already exists %v", requestedSA.Name)
		}
	}

	for _, existingSA := range *existingManagedServiceAccounts {
		if reflect.DeepEqual(existingSA.OwnerReferences, *ownerReferences) {
			matchingRequest := false
			for _, matchingSA := range matchingServiceAccounts {
				if saMatches(&existingSA, &matchingSA) {
					matchingRequest = true
					break
				}
			}

			if !matchingRequest {
				logrus.Infof("Deleting Service Account %v", existingSA.Name)
				err := bc.kubernetesClientSet.CoreV1().ServiceAccounts(existingSA.Namespace).Delete(existingSA.Name, &metav1.DeleteOptions{})
				if err != nil {
					logrus.Infof("Error deleting Service Account: %v", err)
				}
			} else {
				logrus.Debugf("Matches requested Service Account %v", existingSA.Name)
			}
		}
	}

	for _, serviceAccountToCreate := range serviceAccountsToCreate {
		logrus.Infof("Creating Service Account: %v", serviceAccountToCreate.Name)
		_, err := bc.kubernetesClientSet.CoreV1().ServiceAccounts(serviceAccountToCreate.ObjectMeta.Namespace).Create(&serviceAccountToCreate)
		if err != nil {
			logrus.Errorf("Error creating Service Account: %v", err)
		}
	}
}
