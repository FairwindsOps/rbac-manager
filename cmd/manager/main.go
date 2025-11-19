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
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	ctrl "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/fairwindsops/rbac-manager/pkg/apis"
	"github.com/fairwindsops/rbac-manager/pkg/controller"
	"github.com/fairwindsops/rbac-manager/pkg/metrics"
	"github.com/fairwindsops/rbac-manager/pkg/watcher"
	"github.com/fairwindsops/rbac-manager/version"
)

var logLevel = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
var addr = flag.String("metrics-address", ":8042", "The address to serve prometheus metrics.")

func init() {
	klog.InitFlags(nil)
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// slogToLogrAdapter adapts slog.Logger to logr.Logger for controller-runtime
type slogToLogrAdapter struct {
	logger *slog.Logger
}

func (a *slogToLogrAdapter) Init(info logr.RuntimeInfo) {}
func (a *slogToLogrAdapter) Enabled(level int) bool     { return true }
func (a *slogToLogrAdapter) Info(level int, msg string, keysAndValues ...interface{}) {
	a.logger.Info(msg, keysAndValues...)
}
func (a *slogToLogrAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	args := append([]interface{}{"error", err}, keysAndValues...)
	a.logger.Error(msg, args...)
}
func (a *slogToLogrAdapter) WithValues(keysAndValues ...interface{}) logr.LogSink {
	return &slogToLogrAdapter{logger: a.logger.With(keysAndValues...)}
}
func (a *slogToLogrAdapter) WithName(name string) logr.LogSink {
	return &slogToLogrAdapter{logger: a.logger.With("name", name)}
}

func main() {
	flag.Parse()

	level := parseLogLevel(*logLevel)
	opts := &slog.HandlerOptions{
		Level: level,
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	// Set up controller-runtime logger to use slog via adapter
	logrLogger := logr.New(&slogToLogrAdapter{logger: logger})
	ctrl.SetLogger(logrLogger)

	slog.Info("----------------------------------")
	slog.Info("rbac-manager running", "version", version.Version)
	slog.Info("----------------------------------")

	// Get a config to talk to the apiserver
	slog.Debug("Setting up client for manager")
	cfg, err := config.GetConfig()
	if err != nil {
		slog.Error("unable to set up client config", "error", err)
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	slog.Debug("Setting up manager")
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		slog.Error("unable to set up overall controller manager", "error", err)
		os.Exit(1)
	}

	slog.Info("Registering components")

	// Setup Scheme for all resources
	slog.Debug("Setting up scheme")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		slog.Error("unable add APIs to scheme", "error", err)
		os.Exit(1)
	}

	// Setup all Controllers
	slog.Debug("Setting up controller")
	if err := controller.Add(mgr); err != nil {
		slog.Error("unable to register controller to the manager", "error", err)
		os.Exit(1)
	}

	// Watch Related Resources
	slog.Info("Watching resources related to RBAC Definitions")
	watcher.WatchRelatedResources()

	// Start metrics endpoint
	go func() {
		metrics.RegisterMetrics()
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(*addr, nil); err != nil {
			slog.Error("unable to serve the metrics endpoint", "error", err)
			os.Exit(1)
		}
	}()

	// Start the Cmd
	slog.Info("Watching RBAC Definitions")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		slog.Error("unable to run the manager", "error", err)
		os.Exit(1)
	}
}
