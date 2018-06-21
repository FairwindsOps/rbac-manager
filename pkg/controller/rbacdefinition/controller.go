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
	"fmt"
	"reflect"

	"github.com/kubernetes-sigs/kubebuilder/pkg/controller"
	"github.com/kubernetes-sigs/kubebuilder/pkg/controller/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	"k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime/schema"

	logrus "github.com/sirupsen/logrus"

	rbacmanagerv1beta1 "github.com/reactiveops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
	rbacmanagerv1beta1client "github.com/reactiveops/rbac-manager/pkg/client/clientset/versioned/typed/rbacmanager/v1beta1"
	rbacmanagerv1beta1informer "github.com/reactiveops/rbac-manager/pkg/client/informers/externalversions/rbacmanager/v1beta1"
	rbacmanagerv1beta1lister "github.com/reactiveops/rbac-manager/pkg/client/listers/rbacmanager/v1beta1"

	"github.com/reactiveops/rbac-manager/pkg/inject/args"
)

// EDIT THIS FILE
// This files was created by "kubebuilder create resource" for you to edit.
// Controller implementation logic for RBACDefinition resources goes here.

func (bc *RBACDefinitionController) Reconcile(k types.ReconcileKey) error {
	rbacDef, err := bc.rbacDefinitionClient.RBACDefinitions().Get(k.Name, metav1.GetOptions{})

	if err != nil {
		return err
	}

	listOptions := metav1.ListOptions{LabelSelector: "rbac-manager=reactiveops"}
	labels := map[string]string{
		"rbac-manager": "reactiveops",
	}

	ownerReferences := []metav1.OwnerReference{
		*metav1.NewControllerRef(rbacDef, schema.GroupVersionKind{
			Group:   rbacmanagerv1beta1.SchemeGroupVersion.Group,
			Version: rbacmanagerv1beta1.SchemeGroupVersion.Version,
			Kind:    "RBACDefinition",
		}),
	}

	existingManagedClusterRoleBindings, err := bc.kubernetesClientSet.RbacV1().ClusterRoleBindings().List(listOptions)
	existingManagedRoleBindings, err := bc.kubernetesClientSet.RbacV1().RoleBindings("").List(listOptions)
	existingManagedServiceAccounts, err := bc.kubernetesClientSet.CoreV1().ServiceAccounts("").List(listOptions)

	logrus.Infof("RBACDefinition %v\n", rbacDef)

	if rbacDef.RBACBindings == nil {
		logrus.Warn("No RbacBindings defined")
		return nil
	}

	requestedClusterRoleBindings := []rbacv1.ClusterRoleBinding{}
	requestedRoleBindings := []rbacv1.RoleBinding{}
	requestedServiceAccounts := []v1.ServiceAccount{}

	for _, rbacBinding := range rbacDef.RBACBindings {
		for _, requestedSubject := range rbacBinding.Subjects {
			if requestedSubject.Kind == "ServiceAccount" {
				requestedServiceAccounts = append(requestedServiceAccounts, v1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Name:            requestedSubject.Name,
						Namespace:       requestedSubject.Namespace,
						OwnerReferences: ownerReferences,
						Labels:          labels,
					},
				})
			}
		}

		if rbacBinding.ClusterRoleBindings != nil {
			for _, requestedCRB := range rbacBinding.ClusterRoleBindings {
				crbName := fmt.Sprintf("%v-%v-%v", rbacDef.Name, rbacBinding.Name, requestedCRB.ClusterRole)

				logrus.Infof("Processing CRB %v")

				requestedClusterRoleBindings = append(requestedClusterRoleBindings, rbacv1.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:            crbName,
						OwnerReferences: ownerReferences,
						Labels:          labels,
					},
					RoleRef: rbacv1.RoleRef{
						Kind: "ClusterRole",
						Name: requestedCRB.ClusterRole,
					},
					Subjects: rbacBinding.Subjects,
				})
			}
		}

		if rbacBinding.RoleBindings != nil {
			for _, requestedRB := range rbacBinding.RoleBindings {
				objectMeta := metav1.ObjectMeta{
					OwnerReferences: ownerReferences,
					Labels:          labels,
				}

				var requestedRoleName string
				var roleRef rbacv1.RoleRef

				if requestedRB.Namespace == "" {
					logrus.Error("Invalid role binding, namespace required")
					break
				}

				objectMeta.Namespace = requestedRB.Namespace

				if requestedRB.ClusterRole != "" {
					logrus.Infof("Processing Requested ClusterRole %v <> %v <> %v", requestedRB.ClusterRole, requestedRB.Namespace, requestedRB)
					requestedRoleName = requestedRB.ClusterRole
					roleRef = rbacv1.RoleRef{
						Kind: "ClusterRole",
						Name: requestedRB.ClusterRole,
					}
				} else if requestedRB.Role != "" {
					logrus.Infof("Processing Requested Role %v <> %v <> %v", requestedRB.Role, requestedRB.Namespace, requestedRB)
					requestedRoleName = fmt.Sprintf("%v-%v", requestedRB.Role, requestedRB.Namespace)
					roleRef = rbacv1.RoleRef{
						Kind: "Role",
						Name: requestedRB.Role,
					}
				} else {
					logrus.Error("Invalid role binding, role or clusterRole required")
					break
				}

				objectMeta.Name = fmt.Sprintf("%v-%v-%v", rbacDef.Name, rbacBinding.Name, requestedRoleName)

				requestedRoleBindings = append(requestedRoleBindings, rbacv1.RoleBinding{
					ObjectMeta: objectMeta,
					RoleRef:    roleRef,
					Subjects:   rbacBinding.Subjects,
				})
			}
		}
	}

	matchingClusterRoleBindings := []rbacv1.ClusterRoleBinding{}

	for _, requestedCRB := range requestedClusterRoleBindings {
		alreadyExists := false
		for _, existingCRB := range existingManagedClusterRoleBindings.Items {
			if crbMatches(&existingCRB, &requestedCRB) {
				alreadyExists = true
				matchingClusterRoleBindings = append(matchingClusterRoleBindings, existingCRB)
				break
			}
		}

		if !alreadyExists {
			logrus.Infof("Attempting to create Cluster Role Binding: %v", requestedCRB)
			_, err := bc.kubernetesClientSet.RbacV1().ClusterRoleBindings().Create(&requestedCRB)
			if err != nil {
				logrus.Errorf("Error creating Cluster Role Binding: %v", err)
			}
		} else {
			logrus.Infof("Cluster Role Binding already exists %v", requestedCRB)
		}
	}

	for _, existingCRB := range existingManagedClusterRoleBindings.Items {
		if reflect.DeepEqual(existingCRB.OwnerReferences, ownerReferences) {
			matchingRequest := false
			for _, requestedCRB := range matchingClusterRoleBindings {
				if crbMatches(&existingCRB, &requestedCRB) {
					matchingRequest = true
					break
				}
			}

			if !matchingRequest {
				logrus.Infof("Attempting to delete Cluster Role Binding: %v", existingCRB)
				err := bc.kubernetesClientSet.RbacV1().ClusterRoleBindings().Delete(existingCRB.Name, &metav1.DeleteOptions{})
				if err != nil {
					logrus.Errorf("Error deleting Cluster Role Binding: %v", err)
				}
			} else {
				logrus.Infof("Matches requested Cluster Role Binding: %v", err)
			}
		}
	}

	matchingRoleBindings := []rbacv1.RoleBinding{}

	for _, requestedRB := range requestedRoleBindings {
		alreadyExists := false
		for _, existingRB := range existingManagedRoleBindings.Items {
			if rbMatches(&existingRB, &requestedRB) {
				alreadyExists = true
				matchingRoleBindings = append(matchingRoleBindings, existingRB)
				break
			}
		}

		if !alreadyExists {
			logrus.Infof("Attempting to create Role Binding: %v", requestedRB)
			_, err := bc.kubernetesClientSet.RbacV1().RoleBindings(requestedRB.ObjectMeta.Namespace).Create(&requestedRB)
			if err != nil {
				logrus.Errorf("Error creating Role Binding: %v", err)
			}
		} else {
			logrus.Infof("Role Binding already exists %v", requestedRB)
		}
	}

	for _, existingRB := range existingManagedRoleBindings.Items {
		if reflect.DeepEqual(existingRB.OwnerReferences, ownerReferences) {
			matchingRequest := false
			for _, requestedRB := range matchingRoleBindings {
				if rbMatches(&existingRB, &requestedRB) {
					matchingRequest = true
					break
				}
			}

			if !matchingRequest {
				logrus.Infof("Attempting to delete Role Binding %v", existingRB)
				err := bc.kubernetesClientSet.RbacV1().ClusterRoleBindings().Delete(existingRB.Name, &metav1.DeleteOptions{})
				if err != nil {
					logrus.Infof("Error deleting Role Binding: %v", err)
				}
			} else {
				logrus.Infof("Matches requested Role Binding %v", existingRB)
			}
		}
	}

	matchingServiceAccounts := []v1.ServiceAccount{}

	for _, requestedSA := range requestedServiceAccounts {
		alreadyExists := false
		for _, existingSA := range matchingServiceAccounts {
			if saMatches(&existingSA, &requestedSA) {
				alreadyExists = true
				matchingServiceAccounts = append(matchingServiceAccounts, existingSA)
				break
			}
		}

		if !alreadyExists {
			logrus.Infof("Attempting to create Service Account: %v", requestedSA)
			_, err := bc.kubernetesClientSet.CoreV1().ServiceAccounts(requestedSA.ObjectMeta.Namespace).Create(&requestedSA)
			if err != nil {
				logrus.Errorf("Error creating Service Account: %v", err)
			}
		} else {
			logrus.Infof("Service Account already exists %v", requestedSA)
		}
	}

	for _, existingSA := range existingManagedServiceAccounts.Items {
		if reflect.DeepEqual(existingSA.OwnerReferences, ownerReferences) {
			matchingRequest := false
			for _, requestedSA := range matchingServiceAccounts {
				if saMatches(&existingSA, &requestedSA) {
					matchingRequest = true
					break
				}
			}

			if !matchingRequest {
				logrus.Infof("Attempting to delete Service Account %v", existingSA)
				err := bc.kubernetesClientSet.RbacV1().ClusterRoleBindings().Delete(existingSA.Name, &metav1.DeleteOptions{})
				if err != nil {
					logrus.Infof("Error deleting Service Account: %v", err)
				}
			} else {
				logrus.Infof("Matches requested Service Account %v", existingSA)
			}
		}
	}

	return nil
}

// +kubebuilder:controller:group=rbacmanager,version=v1beta1,kind=RBACDefinition,resource=rbacdefinitions
type RBACDefinitionController struct {
	// INSERT ADDITIONAL FIELDS HERE
	rbacDefinitionLister rbacmanagerv1beta1lister.RBACDefinitionLister
	rbacDefinitionClient rbacmanagerv1beta1client.RbacmanagerV1beta1Interface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	rbacDefinitionRecorder record.EventRecorder
	kubernetesClientSet    kubernetes.Interface
}

// ProvideController provides a controller that will be run at startup.  Kubebuilder will use codegeneration
// to automatically register this controller in the inject package
func ProvideController(arguments args.InjectArgs) (*controller.GenericController, error) {
	// INSERT INITIALIZATIONS FOR ADDITIONAL FIELDS HERE
	bc := &RBACDefinitionController{
		rbacDefinitionLister: arguments.ControllerManager.GetInformerProvider(&rbacmanagerv1beta1.RBACDefinition{}).(rbacmanagerv1beta1informer.RBACDefinitionInformer).Lister(),

		rbacDefinitionClient:   arguments.Clientset.RbacmanagerV1beta1(),
		rbacDefinitionRecorder: arguments.CreateRecorder("RBACDefinitionController"),

		kubernetesClientSet: arguments.KubernetesClientSet,
	}

	// Create a new controller that will call RBACDefinitionController.Reconcile on changes to RBACDefinitions
	gc := &controller.GenericController{
		Name:             "RBACDefinitionController",
		Reconcile:        bc.Reconcile,
		InformerRegistry: arguments.ControllerManager,
	}
	if err := gc.Watch(&rbacmanagerv1beta1.RBACDefinition{}); err != nil {
		return gc, err
	}

	// IMPORTANT:
	// To watch additional resource types - such as those created by your controller - add gc.Watch* function calls here
	// Watch function calls will transform each object event into a RBACDefinition Key to be reconciled by the controller.
	//
	// **********
	// For any new Watched types, you MUST add the appropriate // +kubebuilder:informer and // +kubebuilder:rbac
	// annotations to the RBACDefinitionController and run "kubebuilder generate.
	// This will generate the code to start the informers and create the RBAC rules needed for running in a cluster.
	// See:
	// https://godoc.org/github.com/kubernetes-sigs/kubebuilder/pkg/gen/controller#example-package
	// **********

	return gc, nil
}
