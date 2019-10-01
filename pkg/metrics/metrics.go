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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "rbacmanager"

var (
	// ErrorCounter is a global counter for errors
	ErrorCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "errors_total",
			Help:      "Number of errors while reconciling",
		})

	// ChangeCounter counts kubernetes events (e.g. create, delete) on objects (e.g. ClusterRoleBinding)
	ChangeCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "changed_total",
			Help:      "Number of times a Kubernetes object is created or deleted by the rbac-manager",
		},
		[]string{"object", "action"},
	)
)

// RegisterMetrics must be called exactly once and registers the prometheus counters as metrics
func RegisterMetrics() {
	prometheus.MustRegister(ErrorCounter)
	prometheus.MustRegister(ChangeCounter)
	prometheus.MustRegister(prometheus.NewGoCollector())
}
