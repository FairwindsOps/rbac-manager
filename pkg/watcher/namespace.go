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
	kube "github.com/reactiveops/rbac-manager/pkg/kube"
	"github.com/reactiveops/rbac-manager/pkg/reconciler"
	"github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
)

func watchNamespaces(clientset *kubernetes.Clientset) {
	watcher, err := clientset.CoreV1().Namespaces().Watch(kube.ListOptions)

	if err != nil {
		logrus.Error(err, "unable to watch Namespaces")
		runtime.HandleError(err)
	}

	ch := watcher.ResultChan()

	for event := range ch {
		ns, ok := event.Object.(*corev1.Namespace)
		if !ok {
			logrus.Error("Could not parse Namespace")
		}

		logrus.Debugf("Reconciling RBACDefinitions after %s event on  %s Namespace", ns.Name, event.Type)
		rbacDefList, err := kube.GetRbacDefinitions()

		for _, rbacDef := range rbacDefList.Items {
			r := reconciler.Reconciler{Clientset: kube.GetClientsetOrDie()}
			err = r.ReconcileNamespaceChange(&rbacDef, ns)
			if err != nil {
				logrus.Error(err, "unable to reconcile namespace change")
				runtime.HandleError(err)
			}
		}
	}
}
