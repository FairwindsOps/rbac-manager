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

	logrus "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (bc *RBACDefinitionController) reconcileClusterRoleBindings(
	requestedClusterRoleBindings *[]rbacv1.ClusterRoleBinding,
	existingManagedClusterRoleBindings *[]rbacv1.ClusterRoleBinding,
	ownerReferences *[]metav1.OwnerReference) {

	matchingClusterRoleBindings := []rbacv1.ClusterRoleBinding{}

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
			logrus.Infof("Attempting to create Cluster Role Binding: %v", requestedCRB.Name)
			_, err := bc.kubernetesClientSet.RbacV1().ClusterRoleBindings().Create(&requestedCRB)
			if err != nil {
				logrus.Errorf("Error creating Cluster Role Binding: %v", err)
			}
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
				logrus.Infof("Attempting to delete Cluster Role Binding: %v", existingCRB.Name)
				err := bc.kubernetesClientSet.RbacV1().ClusterRoleBindings().Delete(existingCRB.Name, &metav1.DeleteOptions{})
				if err != nil {
					logrus.Errorf("Error deleting Cluster Role Binding: %v", err)
				}
			} else {
				logrus.Debugf("Matches requested Cluster Role Binding: %v", existingCRB.Name)
			}
		}
	}
}

func (bc *RBACDefinitionController) reconcileRoleBindings(
	requestedRoleBindings *[]rbacv1.RoleBinding,
	existingManagedRoleBindings *[]rbacv1.RoleBinding,
	ownerReferences *[]metav1.OwnerReference) {

	matchingRoleBindings := []rbacv1.RoleBinding{}

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
			logrus.Infof("Attempting to create Role Binding: %v", requestedRB.Name)
			_, err := bc.kubernetesClientSet.RbacV1().RoleBindings(requestedRB.ObjectMeta.Namespace).Create(&requestedRB)
			if err != nil {
				logrus.Errorf("Error creating Role Binding: %v", err)
			}
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
				logrus.Infof("Attempting to delete Role Binding %v", existingRB.Name)
				err := bc.kubernetesClientSet.RbacV1().RoleBindings(existingRB.Namespace).Delete(existingRB.Name, &metav1.DeleteOptions{})
				if err != nil {
					logrus.Infof("Error deleting Role Binding: %v", err)
				}
			} else {
				logrus.Debugf("Matches requested Role Binding %v", existingRB.Name)
			}
		}
	}
}

func (bc *RBACDefinitionController) reconcileServiceAccounts(
	requestedServiceAccounts *[]v1.ServiceAccount,
	existingManagedServiceAccounts *[]v1.ServiceAccount,
	ownerReferences *[]metav1.OwnerReference) {

	matchingServiceAccounts := []v1.ServiceAccount{}

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
			logrus.Infof("Attempting to create Service Account: %v", requestedSA.Name)
			_, err := bc.kubernetesClientSet.CoreV1().ServiceAccounts(requestedSA.ObjectMeta.Namespace).Create(&requestedSA)
			if err != nil {
				logrus.Errorf("Error creating Service Account: %v", err)
			}
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
				logrus.Infof("Attempting to delete Service Account %v", existingSA.Name)
				err := bc.kubernetesClientSet.CoreV1().ServiceAccounts(existingSA.Namespace).Delete(existingSA.Name, &metav1.DeleteOptions{})
				if err != nil {
					logrus.Infof("Error deleting Service Account: %v", err)
				}
			} else {
				logrus.Debugf("Matches requested Service Account %v", existingSA.Name)
			}
		}
	}
}
