// Copyright 2018 FairwindsOps Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reconciler

import (
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	rbacmanagerv1beta1 "github.com/fairwindsops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
)

func generateOwnerReferences(name string) []metav1.OwnerReference {
	rbacDef := rbacmanagerv1beta1.RBACDefinition{}
	rbacDef.Name = name

	return []metav1.OwnerReference{
		*metav1.NewControllerRef(&rbacDef, schema.GroupVersionKind{
			Group:   rbacmanagerv1beta1.SchemeGroupVersion.Group,
			Version: rbacmanagerv1beta1.SchemeGroupVersion.Version,
			Kind:    "RBACDefinition",
		}),
	}
}
func TestCrbMatches(t *testing.T) {
	subject1 := rbacv1.Subject{Kind: "User", Name: "joe"}
	subject2 := rbacv1.Subject{Kind: "User", Name: "sue"}
	subject3 := rbacv1.Subject{Kind: "ServiceAccount", Name: "ci"}

	labels1 := map[string]string{"rbac-manager": "fairwinds"}
	labels2 := map[string]string{"something": "else"}

	crb1 := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "hello-world",
			OwnerReferences: []metav1.OwnerReference{},
			Labels:          labels1,
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "Admin",
		},
		Subjects: []rbacv1.Subject{subject1, subject2, subject3},
	}

	crb2 := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "hello-other-world",
			OwnerReferences: generateOwnerReferences("foo"),
			Labels:          labels2,
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "Admin",
		},
		Subjects: []rbacv1.Subject{subject1, subject2, subject3},
	}

	crb3 := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "hello-other-world",
			OwnerReferences: generateOwnerReferences("bar"),
			Labels:          labels1,
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "Admin",
		},
		Subjects: []rbacv1.Subject{subject1, subject2, subject3},
	}

	if crbMatches(&crb1, &crb2) {
		t.Fatal("CRB 1 should not match CRB 2")
	}

	if crbMatches(&crb1, &crb3) {
		t.Fatal("CRB 1 should not match CRB 3")
	}

	if crbMatches(&crb2, &crb3) {
		t.Fatal("CRB 2 should not match CRB 3")
	}

	if !crbMatches(&crb1, &crb1) {
		t.Fatal("CRB 1 should match CRB 1")
	}

	if !crbMatches(&crb2, &crb2) {
		t.Fatal("CRB 2 should match CRB 2")
	}

	if !crbMatches(&crb3, &crb3) {
		t.Fatal("CRB 3 should match CRB 3")
	}
}

func TestRbMatches(t *testing.T) {
	subject1 := rbacv1.Subject{Kind: "User", Name: "joe"}
	subject2 := rbacv1.Subject{Kind: "User", Name: "sue"}
	subject3 := rbacv1.Subject{Kind: "ServiceAccount", Name: "ci"}

	labels1 := map[string]string{"rbac-manager": "fairwinds"}
	labels2 := map[string]string{"something": "else"}

	rb1 := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "hello-world",
			Namespace:       "default",
			OwnerReferences: []metav1.OwnerReference{},
			Labels:          labels1,
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "Admin",
		},
		Subjects: []rbacv1.Subject{subject1, subject2, subject3},
	}

	rb2 := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "hello-other-world",
			Namespace:       "default",
			OwnerReferences: generateOwnerReferences("foo"),
			Labels:          labels2,
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "Admin",
		},
		Subjects: []rbacv1.Subject{subject1, subject2, subject3},
	}

	rb3 := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "hello-other-world",
			Namespace:       "sample",
			OwnerReferences: generateOwnerReferences("bar"),
			Labels:          labels1,
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "Admin",
		},
		Subjects: []rbacv1.Subject{subject1, subject2, subject3},
	}

	if rbMatches(&rb1, &rb2) {
		t.Fatal("RB 1 should not match RB 2")
	}

	if rbMatches(&rb1, &rb3) {
		t.Fatal("RB 1 should not match RB 3")
	}

	if rbMatches(&rb2, &rb3) {
		t.Fatal("RB 2 should not match RB 3")
	}

	if !rbMatches(&rb1, &rb1) {
		t.Fatal("RB 1 should match RB 1")
	}

	if !rbMatches(&rb2, &rb2) {
		t.Fatal("RB 2 should match RB 2")
	}

	if !rbMatches(&rb3, &rb3) {
		t.Fatal("RB 3 should match RB 3")
	}
}
