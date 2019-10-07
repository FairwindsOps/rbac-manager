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
	"os"

	"github.com/fairwindsops/rbac-manager/pkg/apis"
	"github.com/fairwindsops/rbac-manager/pkg/controller"
	"github.com/fairwindsops/rbac-manager/pkg/watcher"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	klog "k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

var (
	// version is set during build
	version = "development"
	// commit is set during build
	commit = "n/a"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	klog.Info("---------------------------------------------------------------")
	klog.Infof("rbac-manager - %s (Git: %s) is starting...", version, commit)
	klog.Info("---------------------------------------------------------------")

	// Get a config to talk to the apiserver
	klog.V(5).Info("Setting up client for manager")
	cfg, err := config.GetConfig()
	if err != nil {
		klog.Error(err, "unable to set up client config")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	klog.V(5).Info("Setting up manager")
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		klog.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	klog.Info("Registering components")

	// Setup Scheme for all resources
	klog.V(5).Info("Setting up scheme")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		klog.Error(err, "unable add APIs to scheme")
		os.Exit(1)
	}

	// Setup all Controllers
	klog.V(5).Info("Setting up controller")
	if err := controller.Add(mgr); err != nil {
		klog.Error(err, "unable to register controller to the manager")
		os.Exit(1)
	}

	// Watch Related Resources
	klog.Info("Watching resources related to RBAC Definitions")
	watcher.WatchRelatedResources()

	// Start the Cmd
	klog.Info("Watching RBAC Definitions")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		klog.Error(err, "unable to run the manager")
		os.Exit(1)
	}
}
