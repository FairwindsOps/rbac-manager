package main

import (
	"flag"
	"log"
	"strings"

	// Import auth/gcp to connect to GKE clusters remotely
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	configlib "github.com/kubernetes-sigs/kubebuilder/pkg/config"
	"github.com/kubernetes-sigs/kubebuilder/pkg/inject/run"
	"github.com/kubernetes-sigs/kubebuilder/pkg/install"
	"github.com/kubernetes-sigs/kubebuilder/pkg/signals"
	extensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/reactiveops/rbac-manager/pkg/inject"
	"github.com/reactiveops/rbac-manager/pkg/inject/args"
	"github.com/reactiveops/rbac-manager/version"

	logrus "github.com/sirupsen/logrus"
)

var installCRDs = flag.Bool("install-crds", true, "install the CRDs used by the controller as part of startup")
var logLevel = flag.String("log-level", logrus.InfoLevel.String(), "Logrus log level")

// Controller-manager main.
func main() {
	flag.Parse()

	stopCh := signals.SetupSignalHandler()

	config := configlib.GetConfigOrDie()

	if *installCRDs {
		if err := install.NewInstaller(config).Install(&InstallStrategy{crds: inject.Injector.CRDs}); err != nil {
			log.Fatalf("Could not create CRDs: %v", err)
		}
	}

	logLevel := logrus.InfoLevel

	if parsed, err := logrus.ParseLevel(logLevel.String()); err == nil {
		logLevel = parsed
	} else {
		// This should theoretically never happen assuming the enum flag
		// is constructed correctly because the enum flag will not allow
		//  an invalid value to be set.
		logrus.Errorf("log-level flag has invalid value %s", strings.ToUpper(logLevel.String()))
	}

	logrus.Infof("RBAC Manager %v Running", version.Version)

	// Start the controllers
	if err := inject.RunAll(run.RunArguments{Stop: stopCh}, args.CreateInjectArgs(config)); err != nil {
		log.Fatalf("%v", err)
	}
}

type InstallStrategy struct {
	install.EmptyInstallStrategy
	crds []*extensionsv1beta1.CustomResourceDefinition
}

func (s *InstallStrategy) GetCRDs() []*extensionsv1beta1.CustomResourceDefinition {
	return s.crds
}
