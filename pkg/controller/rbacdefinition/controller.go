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
	"github.com/kubernetes-sigs/kubebuilder/pkg/controller"
	"github.com/kubernetes-sigs/kubebuilder/pkg/controller/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rbacmanagerv1beta1 "github.com/reactiveops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
	rbacmanagerv1beta1client "github.com/reactiveops/rbac-manager/pkg/client/clientset/versioned/typed/rbacmanager/v1beta1"
	rbacmanagerv1beta1informer "github.com/reactiveops/rbac-manager/pkg/client/informers/externalversions/rbacmanager/v1beta1"
	rbacmanagerv1beta1lister "github.com/reactiveops/rbac-manager/pkg/client/listers/rbacmanager/v1beta1"

	"github.com/reactiveops/rbac-manager/pkg/inject/args"
)

// Reconcile achieves the desired stated defined by an RBACDefinition
func (bc *RBACDefinitionController) Reconcile(k types.ReconcileKey) error {
	rbacDef, err := bc.rbacDefinitionClient.RBACDefinitions().Get(k.Name, metav1.GetOptions{})

	if err != nil {
		return err
	}

	return bc.reconcileRbacDef(rbacDef)
}

// +kubebuilder:controller:group=rbacmanager,version=v1beta1,kind=RBACDefinition,resource=rbacdefinitions
type RBACDefinitionController struct {
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

	return gc, nil
}
