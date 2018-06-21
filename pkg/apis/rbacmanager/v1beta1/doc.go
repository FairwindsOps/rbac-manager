// Api versions allow the api contract for a resource to be changed while keeping
// backward compatibility by support multiple concurrent versions
// of the same resource

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:conversion-gen=github.com/reactiveops/rbac-manager/pkg/apis/rbacmanager
// +k8s:defaulter-gen=TypeMeta
// +groupName=rbacmanager.reactiveops.io
package v1beta1 // import "github.com/reactiveops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
