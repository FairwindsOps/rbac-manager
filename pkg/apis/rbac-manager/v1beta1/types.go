package v1beta1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RbacDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []RbacDefinition `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RbacDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              RbacDefinitionSpec   `json:"spec"`
	Status            RbacDefinitionStatus `json:"status,omitempty"`
}

// RbacDefinitionSpec is a specification for a RbacDefinition resource
type RbacDefinitionSpec struct {
	RbacBindings []RbacBinding `json:"rbacBindings"`
}

// RbacDefinitionStatus is a status struct for a RbacDefinition resource
type RbacDefinitionStatus struct {
	// Fill me
}

// RbacBinding is a specification for a RbacBinding resource
type RbacBinding struct {
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
