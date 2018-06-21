package v1beta1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!
// Created by "kubebuilder create resource" for you to implement the RBACDefinition resource schema definition
// as a go struct.
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RBACDefinitionStatus defines the observed state of RBACDefinition
type RBACDefinitionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "kubebuilder generate" to regenerate code after modifying this file
}

// RBACBinding is a specification for a RBACBinding resource
type RBACBinding struct {
	Name                string               `json:"name"`
	Subjects            []rbacv1.Subject     `json:"subjects"`
	ClusterRoleBindings []ClusterRoleBinding `json:"clusterRoleBindings"`
	RoleBindings        []RoleBinding        `json:"roleBindings"`
}

// ClusterRoleBinding is a specification for a ClusterRoleBinding resource
type ClusterRoleBinding struct {
	ClusterRole string `json:"clusterRole"`
}

// RoleBinding is a specification for a RoleBinding resource
type RoleBinding struct {
	ClusterRole string `json:"clusterRole,omitempty"`
	Role        string `json:"role,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// RBACDefinition
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=rbacdefinitions
type RBACDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	RBACBindings      []RBACBinding        `json:"rbacBindings"`
	Status            RBACDefinitionStatus `json:"status,omitempty"`
}
