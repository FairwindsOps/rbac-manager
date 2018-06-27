package inject

import (
	"github.com/kubernetes-sigs/kubebuilder/pkg/inject/run"
	rbacmanagerv1beta1 "github.com/reactiveops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
	rscheme "github.com/reactiveops/rbac-manager/pkg/client/clientset/versioned/scheme"
	"github.com/reactiveops/rbac-manager/pkg/controller/rbacdefinition"
	"github.com/reactiveops/rbac-manager/pkg/inject/args"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
)

func init() {
	rscheme.AddToScheme(scheme.Scheme)

	// Inject Informers
	Inject = append(Inject, func(arguments args.InjectArgs) error {
		Injector.ControllerManager = arguments.ControllerManager

		if err := arguments.ControllerManager.AddInformerProvider(&rbacmanagerv1beta1.RBACDefinition{}, arguments.Informers.Rbacmanager().V1beta1().RBACDefinitions()); err != nil {
			return err
		}

		// Add Kubernetes informers

		if c, err := rbacdefinition.ProvideController(arguments); err != nil {
			return err
		} else {
			arguments.ControllerManager.AddController(c)
		}
		return nil
	})

	// Inject CRDs
	Injector.CRDs = append(Injector.CRDs, &rbacmanagerv1beta1.RBACDefinitionCRD)
	// Inject PolicyRules
	Injector.PolicyRules = append(Injector.PolicyRules, rbacv1.PolicyRule{
		APIGroups: []string{"rbacmanager.reactiveops.io"},
		Resources: []string{"*"},
		Verbs:     []string{"*"},
	})
	// Inject GroupVersions
	Injector.GroupVersions = append(Injector.GroupVersions, schema.GroupVersion{
		Group:   "rbacmanager.reactiveops.io",
		Version: "v1beta1",
	})
	Injector.RunFns = append(Injector.RunFns, func(arguments run.RunArguments) error {
		Injector.ControllerManager.RunInformersAndControllers(arguments)
		return nil
	})
}
