package v1beta1

import (
	sdkK8sutil "github.com/operator-framework/operator-sdk/pkg/util/k8sutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	CRDName        = "rbacdefinitions.rbac-manager.reactiveops.io"
	ResourceKind   = "RbacDefinition"
	ResourcePlural = "RbacDefinitions"
	GroupName      = "rbac-manager.reactiveops.io"
	ShortName      = "rbacdefinitions"
	Version        = "v1beta1"
)

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
	// SchemeGroupVersion is the group version used to register these objects.
	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}
)

func init() {
	sdkK8sutil.AddToSDKScheme(AddToScheme)
}

// addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&RbacDefinition{},
		&RbacDefinitionList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}