/*
Copyright 2018 FairwindsOps Inc.

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

package controller

import (
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	rbacmanagerv1beta1 "github.com/fairwindsops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
)

// Add creates a new RBACDefinition Controller and adds it to the Manager.
// The Manager will set fields on the Controller and Start it.
func Add(mgr manager.Manager) error {
	var err error

	rbacDef := &rbacmanagerv1beta1.RBACDefinition{}
	err = addController(mgr, newRbacDefReconciler(mgr), "rbacdefinition", rbacDef)

	if err != nil {
		logrus.Errorf("Error adding RBAC Definition reconciler")
		return err
	}

	namespace := &corev1.Namespace{}
	err = addController(mgr, newNamespaceReconciler(mgr), "namespace", namespace)

	if err != nil {
		logrus.Errorf("Error adding Namespace reconciler")
		return err
	}

	return nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func addController(mgr manager.Manager, r reconcile.Reconciler, name string, cType runtime.Object) error {
	// Create a new controller
	c, err := controller.New(name, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Resource
	err = c.Watch(&source.Kind{
		Type: cType,
	}, &handler.EnqueueRequestForObject{})

	if err != nil {
		return err
	}

	return nil
}
