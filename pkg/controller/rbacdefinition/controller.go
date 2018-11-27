/*
Copyright 2018 ReactiveOps.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rbacdefinition

import (
	"context"

	rbacmanagerv1beta1 "github.com/reactiveops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new RBACDefinition Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this rbacmanager.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRBACDefinition{Client: mgr.GetClient(), config: mgr.GetConfig(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rbacdefinition-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to RBACDefinition
	err = c.Watch(&source.Kind{Type: &rbacmanagerv1beta1.RBACDefinition{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by RBACDefinition - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rbacmanagerv1beta1.RBACDefinition{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileRBACDefinition{}

// ReconcileRBACDefinition reconciles a RBACDefinition object
type ReconcileRBACDefinition struct {
	client.Client
	scheme *runtime.Scheme
	config *rest.Config
}

// Reconcile reads that state of the cluster for a RBACDefinition object and makes changes based on the state read
// and what is in the RBACDefinition.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbacmanager.reactiveops.io,resources=rbacdefinitions,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileRBACDefinition) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	var err error
	rdr := Reconciler{}

	// Full Kubernetes ClientSet is required because RBAC types don't
	//   implement methods required for Kubebuilder methods to work
	rdr.k8sClientSet, err = kubernetes.NewForConfig(r.config)

	if err != nil {
		return reconcile.Result{}, err
	}

	// Fetch the RBACDefinition instance
	rbacDef := &rbacmanagerv1beta1.RBACDefinition{}
	err = r.Get(context.TODO(), request.NamespacedName, rbacDef)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	rdr.reconcile(rbacDef)

	return reconcile.Result{}, nil
}
