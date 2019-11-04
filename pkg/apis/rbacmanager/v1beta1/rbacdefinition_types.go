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

package v1beta1

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Subject is an expansion on the rbacv1.Subject to allow definition of ImagePullSecrets for a Service Account
type Subject struct {
	rbacv1.Subject
	ImagePullSecrets []string `json:"imagePullSecrets"`
}

// RBACBinding is a specification for a RBACBinding resource
type RBACBinding struct {
	Name                string               `json:"name"`
	Subjects            []Subject            `json:"subjects"`
	ClusterRoleBindings []ClusterRoleBinding `json:"clusterRoleBindings"`
	RoleBindings        []RoleBinding        `json:"roleBindings"`
}

// ClusterRoleBinding is a specification for a ClusterRoleBinding resource
type ClusterRoleBinding struct {
	ClusterRole string `json:"clusterRole"`
}

// RoleBinding is a specification for a RoleBinding resource
type RoleBinding struct {
	ClusterRole       string               `json:"clusterRole,omitempty"`
	Role              string               `json:"role,omitempty"`
	Namespace         string               `json:"namespace,omitempty"`
	NamespaceSelector metav1.LabelSelector `json:"namespaceSelector,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RBACDefinition is the Schema for the rbacdefinitions API
// +k8s:openapi-gen=true
type RBACDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	RBACBindings      []RBACBinding        `json:"rbacBindings"`
	Status            RBACDefinitionStatus `json:"status,omitempty"`
}

// RBACDefinitionStatus defines the observed state of RBACDefinition
type RBACDefinitionStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RBACDefinitionList contains a list of RBACDefinition
type RBACDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RBACDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RBACDefinition{}, &RBACDefinitionList{})
}
