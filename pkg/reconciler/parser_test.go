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

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	rbacmanagerv1beta1 "github.com/fairwindsops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
)

func TestParseEmpty(t *testing.T) {
	client := fake.NewSimpleClientset()
	testEmpty(t, client, "empty-example")
}

func TestParseStandard(t *testing.T) {
	client := fake.NewSimpleClientset()
	rbacDef := rbacmanagerv1beta1.RBACDefinition{}
	rbacDef.Name = "rbac-config"

	createNamespace(t, client, "web", map[string]string{"app": "web", "team": "devs"})
	createNamespace(t, client, "api", map[string]string{"app": "api", "team": "devs"})
	createNamespace(t, client, "db", map[string]string{"app": "db", "team": "db"})

	rbacDef.RBACBindings = []rbacmanagerv1beta1.RBACBinding{{
		Name: "ci-bot",
		Subjects: []rbacmanagerv1beta1.Subject{{
			Subject: rbacv1.Subject{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      "ci-bot",
				Namespace: "bots",
			},
		}},
		RoleBindings: []rbacmanagerv1beta1.RoleBinding{{
			Namespace: "bots",
			Role:      "custom",
		}},
	}, {
		Name: "devs",
		Subjects: []rbacmanagerv1beta1.Subject{{
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "joe",
			},
		}, {
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "sue",
			},
		}},
		RoleBindings: []rbacmanagerv1beta1.RoleBinding{{
			NamespaceSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"team": "devs"},
			},
			ClusterRole: "edit",
		}},
	}}

	newParseTest(t, client, rbacDef, []rbacv1.RoleBinding{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rbac-config-ci-bot-custom-bots",
			Namespace: "bots",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "Role",
			Name: "custom",
		},
		Subjects: []rbacv1.Subject{{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      "ci-bot",
			Namespace: "bots",
		}},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rbac-config-devs-edit",
			Namespace: "web",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "edit",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "joe",
		}, {
			Kind: rbacv1.UserKind,
			Name: "sue",
		}},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rbac-config-devs-edit",
			Namespace: "api",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "edit",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "joe",
		}, {
			Kind: rbacv1.UserKind,
			Name: "sue",
		}},
	}}, []rbacv1.ClusterRoleBinding{}, []corev1.ServiceAccount{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ci-bot",
			Namespace: "bots",
		},
	}})
}

func TestParseLabels(t *testing.T) {
	client := fake.NewSimpleClientset()
	rbacDef := rbacmanagerv1beta1.RBACDefinition{}
	rbacDef.Name = "rbac-config"

	createNamespace(t, client, "web", map[string]string{"app": "web", "team": "devs"})
	createNamespace(t, client, "api", map[string]string{"app": "api", "team": "devs"})

	rbacDef.RBACBindings = []rbacmanagerv1beta1.RBACBinding{{
		Name: "devs",
		Subjects: []rbacmanagerv1beta1.Subject{{
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "sue",
			},
		}},
		RoleBindings: []rbacmanagerv1beta1.RoleBinding{{
			NamespaceSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"team": "devs"},
			},
			ClusterRole: "edit",
		}},
	}}

	// api and web edit access
	newParseTest(t, client, rbacDef, []rbacv1.RoleBinding{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rbac-config-devs-edit",
			Namespace: "web",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "edit",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "sue",
		}},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rbac-config-devs-edit",
			Namespace: "api",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "edit",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "sue",
		}},
	}}, []rbacv1.ClusterRoleBinding{}, []corev1.ServiceAccount{})

	rbacDef.RBACBindings[0].RoleBindings[0].NamespaceSelector = metav1.LabelSelector{
		MatchLabels: map[string]string{"team": "devs"},
		MatchExpressions: []metav1.LabelSelectorRequirement{{
			Key:      "app",
			Operator: metav1.LabelSelectorOpIn,
			Values:   []string{"web", "queue"},
		}},
	}

	// api edit access
	newParseTest(t, client, rbacDef, []rbacv1.RoleBinding{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rbac-config-devs-edit",
			Namespace: "web",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "edit",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "sue",
		}},
	}}, []rbacv1.ClusterRoleBinding{}, []corev1.ServiceAccount{})

	rbacDef.RBACBindings[0].RoleBindings[0].NamespaceSelector = metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{{
			Key:      "app",
			Operator: metav1.LabelSelectorOpNotIn,
			Values:   []string{"api", "queue"},
		}},
	}

	// web edit access
	newParseTest(t, client, rbacDef, []rbacv1.RoleBinding{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rbac-config-devs-edit",
			Namespace: "web",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "edit",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "sue",
		}},
	}}, []rbacv1.ClusterRoleBinding{}, []corev1.ServiceAccount{})

}

func TestParseMissingNamespace(t *testing.T) {
	client := fake.NewSimpleClientset()
	rbacDef := rbacmanagerv1beta1.RBACDefinition{}
	rbacDef.Name = "rbac-config"

	createNamespace(t, client, "web", map[string]string{"app": "web", "team": "devs"})
	createNamespace(t, client, "api", map[string]string{"app": "api", "team": "devs"})
	createNamespace(t, client, "db", map[string]string{"app": "db", "team": "db"})

	rbacDef.RBACBindings = []rbacmanagerv1beta1.RBACBinding{{
		Name: "devs",
		Subjects: []rbacmanagerv1beta1.Subject{
			{
				Subject: rbacv1.Subject{
					Kind: rbacv1.UserKind,
					Name: "joe",
				},
			},
			{
				Subject: rbacv1.Subject{
					Kind: rbacv1.UserKind,
					Name: "sue",
				},
			},
		},
		RoleBindings: []rbacmanagerv1beta1.RoleBinding{{
			NamespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"team": "other-devs"}},
			ClusterRole:       "edit",
		}},
	}}

	newParseTest(t, client, rbacDef, []rbacv1.RoleBinding{}, []rbacv1.ClusterRoleBinding{}, []corev1.ServiceAccount{})
}

func TestParseMissingSubjects(t *testing.T) {
	client := fake.NewSimpleClientset()
	rbacDef := rbacmanagerv1beta1.RBACDefinition{}
	rbacDef.Name = "rbac-config"

	createNamespace(t, client, "web", map[string]string{"app": "web", "team": "devs"})
	createNamespace(t, client, "api", map[string]string{"app": "api", "team": "devs"})
	createNamespace(t, client, "db", map[string]string{"app": "db", "team": "db"})

	rbacDef.RBACBindings = []rbacmanagerv1beta1.RBACBinding{{
		Name:     "devs",
		Subjects: []rbacmanagerv1beta1.Subject{},
		RoleBindings: []rbacmanagerv1beta1.RoleBinding{{
			NamespaceSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{"team": "devs"},
			},
			ClusterRole: "edit",
		}},
	}}

	newParseTest(t, client, rbacDef, []rbacv1.RoleBinding{}, []rbacv1.ClusterRoleBinding{}, []corev1.ServiceAccount{})
}

func TestManagerToRbacSubjects(t *testing.T) {
	expected := []rbacv1.Subject{
		{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      "robot",
			Namespace: "default",
		},
	}
	subjects := []rbacmanagerv1beta1.Subject{
		{
			Subject: expected[0],
		},
	}
	actual := managerSubjectsToRbacSubjects(subjects)
	assert.ElementsMatch(t, expected, actual, "expected subjects to match")
}

func newParseTest(t *testing.T, client *fake.Clientset, rbacDef rbacmanagerv1beta1.RBACDefinition, expectedRb []rbacv1.RoleBinding, expectedCrb []rbacv1.ClusterRoleBinding, expectedSa []corev1.ServiceAccount) {
	p := Parser{Clientset: client}

	err := p.Parse(rbacDef)
	if err != nil {
		t.Logf("Error parsing RBAC Definition: %v", err)
	}

	expectParsedRB(t, p, expectedRb)
	expectParsedCRB(t, p, expectedCrb)
	expectParsedSA(t, p, expectedSa)
}

func testEmpty(t *testing.T, client *fake.Clientset, name string) {
	rbacDef := rbacmanagerv1beta1.RBACDefinition{}
	rbacDef.Name = name
	rbacDef.RBACBindings = []rbacmanagerv1beta1.RBACBinding{}

	newParseTest(t, client, rbacDef, []rbacv1.RoleBinding{}, []rbacv1.ClusterRoleBinding{}, []corev1.ServiceAccount{})
}

func expectParsedCRB(t *testing.T, p Parser, expected []rbacv1.ClusterRoleBinding) {
	assert.Len(t, p.parsedClusterRoleBindings, len(expected), "Expected length to match")

	for _, expectedCrb := range expected {
		matchFound := false
		for _, actualCrb := range p.parsedClusterRoleBindings {
			if actualCrb.Name == expectedCrb.Name {
				matchFound = true
				assert.ElementsMatch(t, expectedCrb.Subjects, actualCrb.Subjects, "Expected subjects to match")
				assert.EqualValues(t, expectedCrb.RoleRef, actualCrb.RoleRef, "Expected role ref to match")
				break
			}
		}

		if !matchFound {
			t.Fatalf("Matching cluster role binding not found for %v", expectedCrb.Name)
		}
	}
}

func expectParsedRB(t *testing.T, p Parser, expected []rbacv1.RoleBinding) {
	assert.Len(t, p.parsedRoleBindings, len(expected), "Expected length to match")

	for _, expectedRb := range expected {
		matchFound := false
		for _, actualRb := range p.parsedRoleBindings {
			if actualRb.Name == expectedRb.Name && expectedRb.Namespace == actualRb.Namespace {
				matchFound = true
				assert.ElementsMatch(t, expectedRb.Subjects, actualRb.Subjects, "Expected subjects to match")
				assert.EqualValues(t, expectedRb.RoleRef, actualRb.RoleRef, "Expected role ref to match")
				break
			}
		}

		if !matchFound {
			t.Fatalf("Matching role binding not found for %v", expectedRb.Name)
		}
	}
}

func expectParsedSA(t *testing.T, p Parser, expected []corev1.ServiceAccount) {
	assert.Len(t, p.parsedServiceAccounts, len(expected), "Expected length to match")

	for _, expectedSa := range expected {
		matchFound := false
		for _, actualSa := range p.parsedServiceAccounts {
			if actualSa.Name == expectedSa.Name && expectedSa.Namespace == actualSa.Namespace {
				matchFound = true
				break
			}
		}

		if !matchFound {
			t.Fatalf("Matching service account not found for %v", expectedSa.Name)
		}
	}
}

func createNamespace(t *testing.T, client *fake.Clientset, name string, labels map[string]string) {
	_, err := client.CoreV1().Namespaces().Create(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: labels,
			},
		},
	)

	if err != nil {
		t.Fatalf("Error creating namespace %v", err)
	}
}
