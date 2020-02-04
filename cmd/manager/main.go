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

package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"github.com/fairwindsops/rbac-manager/pkg/apis"
	"github.com/fairwindsops/rbac-manager/pkg/controller"
	"github.com/fairwindsops/rbac-manager/pkg/metrics"
	"github.com/fairwindsops/rbac-manager/pkg/watcher"
	"github.com/fairwindsops/rbac-manager/version"
)

var logLevel = flag.String("log-level", logrus.InfoLevel.String(), "Logrus log level")
var addr = flag.String("metrics-address", ":8080", "The address to serve prometheus metrics.")

func init() {
	klog.InitFlags(nil)
}

func main() {
	flag.Parse()

	parsedLevel, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		// This should theoretically never happen
		logrus.Errorf("log-level flag has invalid value %s", *logLevel)
	} else {
		logrus.SetLevel(parsedLevel)
	}

	logrus.Info("----------------------------------")
	logrus.Infof("rbac-manager %v running", version.Version)
	logrus.Info("----------------------------------")

	// Get a config to talk to the apiserver
	logrus.Debug("Setting up client for manager")
	cfg, err := config.GetConfig()
	if err != nil {
		logrus.Error(err, "unable to set up client config")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	logrus.Debug("Setting up manager")
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		logrus.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	logrus.Info("Registering components")

	// Setup Scheme for all resources
	logrus.Debug("Setting up scheme")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		logrus.Error(err, "unable add APIs to scheme")
		os.Exit(1)
	}

	// Setup all Controllers
	logrus.Debug("Setting up controller")
	if err := controller.Add(mgr); err != nil {
		logrus.Error(err, "unable to register controller to the manager")
		os.Exit(1)
	}

	// Watch Related Resources
	logrus.Info("Watching resources related to RBAC Definitions")
	watcher.WatchRelatedResources()

	// Start metrics endpoint
	go func() {
		metrics.RegisterMetrics()
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(*addr, nil); err != nil {
			logrus.Error(err, "unable to serve the metrics endpoint")
			os.Exit(1)
		}
	}()

	// Start the Cmd
	logrus.Info("Watching RBAC Definitions")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logrus.Error(err, "unable to run the manager")
		os.Exit(1)
	}
}
