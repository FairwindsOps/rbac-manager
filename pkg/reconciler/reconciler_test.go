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
	"github.com/fairwindsops/rbac-manager/pkg/kube"
)

func TestReconcileRbacDefEmpty(t *testing.T) {
	client := fake.NewSimpleClientset()
	rbacDef := rbacmanagerv1beta1.RBACDefinition{}
	rbacDef.Name = "empty-example"
	testEmptyExample(t, client, rbacDef.Name)
}

func TestReconcileRbacDefChanges(t *testing.T) {
	client := fake.NewSimpleClientset()
	rbacDef := rbacmanagerv1beta1.RBACDefinition{}
	rbacDef.Name = "changing-example"
	testEmptyExample(t, client, rbacDef.Name)

	rbacDef.RBACBindings = []rbacmanagerv1beta1.RBACBinding{{
		Name: "admins",
		Subjects: []rbacmanagerv1beta1.Subject{{
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "jan",
			},
		}},
		ClusterRoleBindings: []rbacmanagerv1beta1.ClusterRoleBinding{{
			ClusterRole: "admin",
		}},
		RoleBindings: []rbacmanagerv1beta1.RoleBinding{},
	}}

	newReconcileTest(t, client, rbacDef, []rbacv1.RoleBinding{}, []rbacv1.ClusterRoleBinding{{
		ObjectMeta: metav1.ObjectMeta{
			Name: "changing-example-admins-admin",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "admin",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "jan",
		}},
	}}, []corev1.ServiceAccount{})

	rbacDef.RBACBindings = []rbacmanagerv1beta1.RBACBinding{{
		Name: "admins",
		Subjects: []rbacmanagerv1beta1.Subject{{
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "jan",
			}}, {
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "joe",
			},
		}},
		ClusterRoleBindings: []rbacmanagerv1beta1.ClusterRoleBinding{{
			ClusterRole: "cluster-admin",
		}},
		RoleBindings: []rbacmanagerv1beta1.RoleBinding{},
	}}

	newReconcileTest(t, client, rbacDef, []rbacv1.RoleBinding{}, []rbacv1.ClusterRoleBinding{{
		ObjectMeta: metav1.ObjectMeta{
			Name: "changing-example-admins-cluster-admin",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "cluster-admin",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "jan",
		}, {
			Kind: rbacv1.UserKind,
			Name: "joe",
		}},
	}}, []corev1.ServiceAccount{})

	testEmptyExample(t, client, rbacDef.Name)
}

func TestReconcileRbacDefServiceAccounts(t *testing.T) {
	client := fake.NewSimpleClientset()
	rbacDef := rbacmanagerv1beta1.RBACDefinition{}
	rbacDef.Name = "service-account-example"
	testEmptyExample(t, client, rbacDef.Name)

	rbacDef.RBACBindings = []rbacmanagerv1beta1.RBACBinding{{
		Name: "ci-bot",
		Subjects: []rbacmanagerv1beta1.Subject{{
			Subject: rbacv1.Subject{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      "ci-bot",
				Namespace: "bots",
			},
		}},
		ClusterRoleBindings: []rbacmanagerv1beta1.ClusterRoleBinding{},
		RoleBindings: []rbacmanagerv1beta1.RoleBinding{{
			Namespace:   "web",
			ClusterRole: "view",
		}, {
			Namespace: "bots",
			Role:      "custom",
		}},
	}}

	newReconcileTest(t, client, rbacDef, []rbacv1.RoleBinding{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-account-example-ci-bot-view",
			Namespace: "web",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "view",
		},
		Subjects: []rbacv1.Subject{{
			Kind:      rbacv1.ServiceAccountKind,
			Name:      "ci-bot",
			Namespace: "bots",
		}},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-account-example-ci-bot-custom-bots",
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
	}}, []rbacv1.ClusterRoleBinding{}, []corev1.ServiceAccount{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ci-bot",
			Namespace: "bots",
		},
	}})

	testEmptyExample(t, client, rbacDef.Name)
}

func TestReconcileRbacDefInvalid(t *testing.T) {
	client := fake.NewSimpleClientset()
	rbacDef := rbacmanagerv1beta1.RBACDefinition{}
	rbacDef.Name = "invalid-example"
	testEmptyExample(t, client, rbacDef.Name)

	// missing namespace in RoleBinding
	rbacDef.RBACBindings = []rbacmanagerv1beta1.RBACBinding{{
		Name: "ci-bot",
		Subjects: []rbacmanagerv1beta1.Subject{{
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "joe",
			},
		}},
		ClusterRoleBindings: []rbacmanagerv1beta1.ClusterRoleBinding{},
		RoleBindings: []rbacmanagerv1beta1.RoleBinding{{
			ClusterRole: "view",
		}},
	}}

	// nothing should get created
	newReconcileTest(t, client, rbacDef, []rbacv1.RoleBinding{}, []rbacv1.ClusterRoleBinding{}, []corev1.ServiceAccount{})
}

func TestReconcileNamespaceChangesLabels(t *testing.T) {
	var err error

	client := fake.NewSimpleClientset()
	rbacDefMatchLabels := rbacmanagerv1beta1.RBACDefinition{}
	rbacDefMatchLabels.Name = "namespace-selector-match-labels"

	_, err = client.CoreV1().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "web",
			Labels: map[string]string{"team": "dev", "app": "web"},
		},
	})
	if err != nil {
		t.Fatalf("Error creating namespace %#v", err)
	}

	_, err = client.CoreV1().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "api",
			Labels: map[string]string{"team": "dev", "app": "api"},
		},
	})
	if err != nil {
		t.Fatalf("Error creating namespace %#v", err)
	}

	testEmptyExample(t, client, rbacDefMatchLabels.Name)

	// Match Labels rbacdef
	rbacDefMatchLabels.RBACBindings = []rbacmanagerv1beta1.RBACBinding{{
		Name: "web-app",
		Subjects: []rbacmanagerv1beta1.Subject{{
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "Joe",
			},
		}, {
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "Sue",
			},
		}},
		RoleBindings: []rbacmanagerv1beta1.RoleBinding{{
			ClusterRole:       "edit",
			NamespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"app": "web"}},
		}},
	}, {
		Name: "dev-team",
		Subjects: []rbacmanagerv1beta1.Subject{{
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "Joe",
			},
		}, {
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "Sue",
			},
		}, {
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "Kay",
			},
		}},
		RoleBindings: []rbacmanagerv1beta1.RoleBinding{{
			ClusterRole:       "view",
			NamespaceSelector: metav1.LabelSelector{MatchLabels: map[string]string{"team": "dev"}},
		}},
	}}

	// Test the matchlabels scenario
	newReconcileNamespaceChangesTest(t, client, rbacDefMatchLabels, []rbacv1.RoleBinding{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "namespace-selector-match-labels-web-app-edit",
			Namespace: "web",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "edit",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "Joe",
		}, {
			Kind: rbacv1.UserKind,
			Name: "Sue",
		}},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "namespace-selector-match-labels-dev-team-view",
			Namespace: "web",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "view",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "Joe",
		}, {
			Kind: rbacv1.UserKind,
			Name: "Sue",
		}, {
			Kind: rbacv1.UserKind,
			Name: "Kay",
		}},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "namespace-selector-match-labels-dev-team-view",
			Namespace: "api",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "view",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "Joe",
		}, {
			Kind: rbacv1.UserKind,
			Name: "Sue",
		}, {
			Kind: rbacv1.UserKind,
			Name: "Kay",
		}},
	}})
}

func TestReconcileNamespaceChangesExpressions(t *testing.T) {
	var err error

	client := fake.NewSimpleClientset()
	rbacDefMatchExpressions := rbacmanagerv1beta1.RBACDefinition{}
	rbacDefMatchExpressions.Name = "namespace-selector-match-expressions"

	_, err = client.CoreV1().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "web",
			Labels: map[string]string{"team": "dev", "app": "web"},
		},
	})
	if err != nil {
		t.Fatalf("Error creating namespace %#v", err)
	}

	_, err = client.CoreV1().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "api",
			Labels: map[string]string{"team": "dev", "app": "api"},
		},
	})
	if err != nil {
		t.Fatalf("Error creating namespace %#v", err)
	}
	testEmptyExample(t, client, rbacDefMatchExpressions.Name)

	// Match Expressions rbacdef
	rbacDefMatchExpressions.RBACBindings = []rbacmanagerv1beta1.RBACBinding{{
		Name: "web-app",
		Subjects: []rbacmanagerv1beta1.Subject{{
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "Joe",
			},
		}, {
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "Sue",
			},
		}},
		RoleBindings: []rbacmanagerv1beta1.RoleBinding{{
			ClusterRole: "edit",
			NamespaceSelector: metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{{
					Key:      "app",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"web"},
				}},
			},
		}},
	}, {
		Name: "dev-team",
		Subjects: []rbacmanagerv1beta1.Subject{{
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "Joe",
			},
		}, {
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "Sue",
			},
		}, {
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "Kay",
			},
		}},
		RoleBindings: []rbacmanagerv1beta1.RoleBinding{{
			ClusterRole: "view",
			NamespaceSelector: metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{{
					Key:      "team",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"dev"},
				}},
			}},
		}}}

	// Test the matchexpressions scenario
	newReconcileNamespaceChangesTest(t, client, rbacDefMatchExpressions, []rbacv1.RoleBinding{{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "namespace-selector-match-expressions-web-app-edit",
			Namespace: "web",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "edit",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "Joe",
		}, {
			Kind: rbacv1.UserKind,
			Name: "Sue",
		}},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "namespace-selector-match-expressions-dev-team-view",
			Namespace: "web",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "view",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "Joe",
		}, {
			Kind: rbacv1.UserKind,
			Name: "Sue",
		}, {
			Kind: rbacv1.UserKind,
			Name: "Kay",
		}},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name:      "namespace-selector-match-expressions-dev-team-view",
			Namespace: "api",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "view",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "Joe",
		}, {
			Kind: rbacv1.UserKind,
			Name: "Sue",
		}, {
			Kind: rbacv1.UserKind,
			Name: "Kay",
		}},
	}})
}

// The namespace reconciler should not reconcile CRBs since they are not namespaced.
func TestReconcileNamespaceChangesCRB(t *testing.T) {
	var err error

	client := fake.NewSimpleClientset()
	rbacDef := rbacmanagerv1beta1.RBACDefinition{}
	rbacDef.Name = "namespace-selector-empty"

	_, err = client.CoreV1().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "web",
			Labels: map[string]string{"team": "dev", "app": "web"},
		},
	})
	if err != nil {
		t.Fatalf("Error creating namespace %#v", err)
	}

	_, err = client.CoreV1().Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "api",
			Labels: map[string]string{"team": "dev", "app": "api"},
		},
	})
	if err != nil {
		t.Fatalf("Error creating namespace %#v", err)
	}

	// Match Expressions rbacdef
	rbacDef.RBACBindings = []rbacmanagerv1beta1.RBACBinding{{
		Name: "web-app",
		Subjects: []rbacmanagerv1beta1.Subject{{
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "Joe",
			},
		}, {
			Subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "Sue",
			},
		}},
		ClusterRoleBindings: []rbacmanagerv1beta1.ClusterRoleBinding{{
			ClusterRole: "edit"},
		}},
	}

	// Test the matchexpressions scenario
	newReconcileNamespaceChangesTest(t, client, rbacDef, []rbacv1.RoleBinding{})
}

func newReconcileTest(t *testing.T, client *fake.Clientset, rbacDef rbacmanagerv1beta1.RBACDefinition, expectedRb []rbacv1.RoleBinding, expectedCrb []rbacv1.ClusterRoleBinding, expectedSa []corev1.ServiceAccount) {
	r := Reconciler{Clientset: client}
	_ = r.Reconcile(&rbacDef)
	expectRoleBindings(t, client, expectedRb)
	expectClusterRoleBindings(t, client, expectedCrb)
	expectServiceAccounts(t, client, expectedSa)
}

func newReconcileNamespaceChangesTest(t *testing.T, client *fake.Clientset, rbacDef rbacmanagerv1beta1.RBACDefinition, expectedRb []rbacv1.RoleBinding) {
	r := Reconciler{Clientset: client}
	// Namespace doesn't matter here, just used for logging
	_ = r.ReconcileNamespaceChange(&rbacDef, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
	})
	expectRoleBindings(t, client, expectedRb)
	expectClusterRoleBindings(t, client, []rbacv1.ClusterRoleBinding{})
}

func testEmptyExample(t *testing.T, client *fake.Clientset, name string) {
	rbacDef := rbacmanagerv1beta1.RBACDefinition{}
	rbacDef.Name = name
	rbacDef.RBACBindings = []rbacmanagerv1beta1.RBACBinding{}

	newReconcileTest(t, client, rbacDef, []rbacv1.RoleBinding{}, []rbacv1.ClusterRoleBinding{}, []corev1.ServiceAccount{})
}

func expectClusterRoleBindings(t *testing.T, client *fake.Clientset, expected []rbacv1.ClusterRoleBinding) {
	actual, err := client.RbacV1().ClusterRoleBindings().List(kube.ListOptions)

	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, actual.Items, len(expected), "Expected length to match")

	for _, expectedCrb := range expected {
		matchFound := false
		for _, actualCrb := range actual.Items {
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

func expectRoleBindings(t *testing.T, client *fake.Clientset, expected []rbacv1.RoleBinding) {
	actual, err := client.RbacV1().RoleBindings("").List(kube.ListOptions)

	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, actual.Items, len(expected), "Expected length to match")

	for _, expectedRb := range expected {
		matchFound := false
		for _, actualRb := range actual.Items {
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

func expectServiceAccounts(t *testing.T, client *fake.Clientset, expected []corev1.ServiceAccount) {
	actual, err := client.CoreV1().ServiceAccounts("").List(metav1.ListOptions{})

	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, actual.Items, len(expected), "Expected length to match")

	for _, expectedSa := range expected {
		matchFound := false
		for _, actualSa := range actual.Items {
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
