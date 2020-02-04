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
	"context"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rbacmanagerv1beta1 "github.com/fairwindsops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
	"github.com/fairwindsops/rbac-manager/pkg/kube"
	"github.com/fairwindsops/rbac-manager/pkg/metrics"
	"github.com/fairwindsops/rbac-manager/pkg/reconciler"
)

// newNamespaceReconciler returns a new reconcile.Reconciler
func newNamespaceReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNamespace{Client: mgr.GetClient(), config: mgr.GetConfig(), scheme: mgr.GetScheme()}
}

// ReconcileNamespace reconciles a Namespace object
type ReconcileNamespace struct {
	client.Client
	scheme *runtime.Scheme
	config *rest.Config
}

// Reconcile makes changes in response to Namespace changes
func (r *ReconcileNamespace) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	var err error

	// Fetch the Namespace
	namespace := &v1.Namespace{}
	err = r.Get(context.TODO(), request.NamespacedName, namespace)

	if err != nil {
		if errors.IsNotFound(err) {
			err = reconcileNamespace(r.config, namespace)
			if err != nil {
				metrics.ErrorCounter.Inc()
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		metrics.ErrorCounter.Inc()
		return reconcile.Result{}, err
	}

	err = reconcileNamespace(r.config, namespace)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func reconcileNamespace(config *rest.Config, namespace *v1.Namespace) error {
	metrics.ReconcileCounter.WithLabelValues("namespace").Inc()
	var err error
	var rbacDefList rbacmanagerv1beta1.RBACDefinitionList
	rdr := reconciler.Reconciler{}

	// Full Kubernetes ClientSet is required because RBAC types don't
	//   implement methods required for controller-runtime methods to work
	rdr.Clientset, err = kubernetes.NewForConfig(config)

	if err != nil {
		return err
	}

	rbacDefList, err = kube.GetRbacDefinitions()
	if err != nil {
		return err
	}

	for _, rbacDef := range rbacDefList.Items {
		err = rdr.ReconcileNamespaceChange(&rbacDef, namespace)
		if err != nil {
			return err
		}
	}

	return nil
}
