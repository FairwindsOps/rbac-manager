package v1beta1

import (
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: "rbacmanager.reactiveops.io", Version: "v1beta1"}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&RBACDefinition{},
		&RBACDefinitionList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RBACDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RBACDefinition `json:"items"`
}

// CRD Generation
func getFloat(f float64) *float64 {
	return &f
}

func getInt(i int64) *int64 {
	return &i
}

var (
	// Define CRDs for resources
	RBACDefinitionCRD = v1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rbacdefinitions.rbacmanager.reactiveops.io",
		},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group:   "rbacmanager.reactiveops.io",
			Version: "v1beta1",
			Names: v1beta1.CustomResourceDefinitionNames{
				Kind:   "RBACDefinition",
				Plural: "rbacdefinitions",
			},
			Scope: "Cluster",
			Validation: &v1beta1.CustomResourceValidation{
				OpenAPIV3Schema: &v1beta1.JSONSchemaProps{
					Type: "object",
					Properties: map[string]v1beta1.JSONSchemaProps{
						"apiVersion": v1beta1.JSONSchemaProps{
							Type: "string",
						},
						"kind": v1beta1.JSONSchemaProps{
							Type: "string",
						},
						"metadata": v1beta1.JSONSchemaProps{
							Type: "object",
						},
						"rbacBindings": v1beta1.JSONSchemaProps{
							Type: "array",
							Items: &v1beta1.JSONSchemaPropsOrArray{
								Schema: &v1beta1.JSONSchemaProps{
									Type: "object",
									Properties: map[string]v1beta1.JSONSchemaProps{
										"clusterRoleBindings": v1beta1.JSONSchemaProps{
											Type: "array",
											Items: &v1beta1.JSONSchemaPropsOrArray{
												Schema: &v1beta1.JSONSchemaProps{
													Type: "object",
													Properties: map[string]v1beta1.JSONSchemaProps{
														"clusterRole": v1beta1.JSONSchemaProps{
															Type: "string",
														},
													},
													Required: []string{
														"clusterRole",
													}},
											},
										},
										"name": v1beta1.JSONSchemaProps{
											Type: "string",
										},
										"roleBindings": v1beta1.JSONSchemaProps{
											Type: "array",
											Items: &v1beta1.JSONSchemaPropsOrArray{
												Schema: &v1beta1.JSONSchemaProps{
													Type: "object",
													Properties: map[string]v1beta1.JSONSchemaProps{
														"clusterRole": v1beta1.JSONSchemaProps{
															Type: "string",
														},
														"namespace": v1beta1.JSONSchemaProps{
															Type: "string",
														},
														"role": v1beta1.JSONSchemaProps{
															Type: "string",
														},
													},
												},
											},
										},
										"subjects": v1beta1.JSONSchemaProps{
											Type: "array",
											Items: &v1beta1.JSONSchemaPropsOrArray{
												Schema: &v1beta1.JSONSchemaProps{
													Type:       "object",
													Properties: map[string]v1beta1.JSONSchemaProps{},
												},
											},
										},
									},
									Required: []string{
										"name",
										"subjects",
										"clusterRoleBindings",
										"roleBindings",
									}},
							},
						},
						"status": v1beta1.JSONSchemaProps{
							Type:       "object",
							Properties: map[string]v1beta1.JSONSchemaProps{},
						},
					},
					Required: []string{
						"metadata",
						"rbacBindings",
					}},
			},
		},
	}
)
