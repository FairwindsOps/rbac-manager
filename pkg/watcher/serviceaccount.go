/*
Copyright 2019 FairwindsOps Inc.

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
	kube "github.com/fairwindsops/rbac-manager/pkg/kube"
	"github.com/fairwindsops/rbac-manager/pkg/reconciler"
	"k8s.io/klog"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

func watchServiceAccounts(clientset *kubernetes.Clientset) {
	watcher, err := clientset.CoreV1().ServiceAccounts("").Watch(kube.ListOptions)

	if err != nil {
		klog.Errorf("Unable to watch Service Accounts: %v", err)
		runtime.HandleError(err)
	}

	ch := watcher.ResultChan()

	for event := range ch {
		sa, ok := event.Object.(*corev1.ServiceAccount)
		if !ok {
			klog.Error("Could not parse Service Account")
		} else if event.Type == watch.Modified || event.Type == watch.Deleted {
			klog.V(5).Infof("Reconciling RBACDefinition for %s ServiceAccount after %s event", sa.Name, event.Type)
			r := reconciler.Reconciler{Clientset: kube.GetClientsetOrDie()}
			r.ReconcileOwners(sa.OwnerReferences, "ServiceAccount")
		}
	}
}
