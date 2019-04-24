/*
Copyright 2019 ReactiveOps.

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

package watcher

import (
	"log"

	kube "github.com/reactiveops/rbac-manager/pkg/kube"
	"github.com/reactiveops/rbac-manager/pkg/reconciler"
	"github.com/sirupsen/logrus"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

func watchRoleBindings(clientset *kubernetes.Clientset) error {
	watcher, err := clientset.RbacV1().RoleBindings("").Watch(kube.ListOptions)

	if err != nil {
		logrus.Error(err, "unable to watch Role Bindings")
		return err
	}

	ch := watcher.ResultChan()

	for event := range ch {
		rb, ok := event.Object.(*rbacv1.RoleBinding)
		if !ok {
			log.Fatal("unexpected type")
		}

		log.Printf("RB %v", rb)

		if event.Type == watch.Modified || event.Type == watch.Deleted {
			reconciler.ReconcileOwners(rb.OwnerReferences, "RoleBinding")
		}
	}
	return nil
}
