package rbacdefinition

import (
	"github.com/stretchr/testify/assert"
	"testing"

	rbacmanagerv1beta1 "github.com/reactiveops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var listOptions = metav1.ListOptions{LabelSelector: "rbac-manager=reactiveops"}

func TestReconcileRbacDef(t *testing.T) {
	client := fake.NewSimpleClientset()
	rdc := RBACDefinitionController{}
	rdc.kubernetesClientSet = client

	client.RbacV1().ClusterRoleBindings().List(listOptions)
	clusterRoleBindings, err := client.RbacV1().ClusterRoleBindings().List(listOptions)

	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, clusterRoleBindings.Items, 0, "More than 0 cluster role bindings to start")

	rbacDef := rbacmanagerv1beta1.RBACDefinition{}
	rbacDef.RBACBindings = []rbacmanagerv1beta1.RBACBinding{{
		Name: "example-1",
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "bar",
		}},
		ClusterRoleBindings: []rbacmanagerv1beta1.ClusterRoleBinding{{
			ClusterRole: "admin",
		}},
		RoleBindings: []rbacmanagerv1beta1.RoleBinding{},
	}}
	rbacDef.Name = "test-config"

	rdc.reconcileRbacDef(&rbacDef)

	expectClusterRoleBindings(t, client, []rbacv1.ClusterRoleBinding{{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-config-example-1-admin",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "admin",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "bar",
		}},
	}})

	rbacDef.RBACBindings = []rbacmanagerv1beta1.RBACBinding{{
		Name: "example-1",
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "bar",
		}},
		ClusterRoleBindings: []rbacmanagerv1beta1.ClusterRoleBinding{{
			ClusterRole: "edit",
		}},
		RoleBindings: []rbacmanagerv1beta1.RoleBinding{},
	}, {
		Name: "example-2",
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "joe",
		}, {
			Kind: rbacv1.UserKind,
			Name: "sue",
		}},
		ClusterRoleBindings: []rbacmanagerv1beta1.ClusterRoleBinding{{
			ClusterRole: "view",
		}},
		RoleBindings: []rbacmanagerv1beta1.RoleBinding{{
			ClusterRole: "admin",
			Namespace:   "web",
		}, {
			ClusterRole: "edit",
			Namespace:   "api",
		}},
	}}
	rdc.reconcileRbacDef(&rbacDef)

	expectClusterRoleBindings(t, client, []rbacv1.ClusterRoleBinding{{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-config-example-1-edit",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "edit",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "bar",
		}},
	}, {
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-config-example-2-view",
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "view",
		},
		Subjects: []rbacv1.Subject{{
			Kind: rbacv1.UserKind,
			Name: "joe",
		}, {
			Kind: rbacv1.UserKind,
			Name: "sue",
		}},
	}})
}

func expectClusterRoleBindings(t *testing.T, client *fake.Clientset, expected []rbacv1.ClusterRoleBinding) {
	actual, err := client.RbacV1().ClusterRoleBindings().List(listOptions)

	if err != nil {
		t.Fatal(err)
	}

	for index, actualCrb := range actual.Items {
		expectedCrb := expected[index]
		assert.Equal(t, expectedCrb.Name, actualCrb.Name, "Expected role ref to match")
		assert.ElementsMatch(t, expectedCrb.Subjects, actualCrb.Subjects, "Expected subjects to match")
		assert.EqualValues(t, expectedCrb.RoleRef, actualCrb.RoleRef, "Expected role ref to match")
	}
}
