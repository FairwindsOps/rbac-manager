package rbacdefinition

import (
	"testing"

	rbacmanagerv1beta1 "github.com/reactiveops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestReconcileRbacDef(t *testing.T) {
	listOptions := metav1.ListOptions{LabelSelector: "rbac-manager=reactiveops"}

	client := fake.NewSimpleClientset()
	client.RbacV1().ClusterRoleBindings().List(listOptions)
	clusterRoleBindings, err := client.RbacV1().ClusterRoleBindings().List(listOptions)

	if err != nil {
		t.Fatal(err)
	}

	if len(clusterRoleBindings.Items) > 0 {
		t.Fatal("More than 0 cluster role bindings to start")
	}

	name := "test-config"
	rbacBindings := []rbacmanagerv1beta1.RBACBinding{{
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

	rbacDef := rbacmanagerv1beta1.RBACDefinition{}
	rbacDef.Name = name
	rbacDef.RBACBindings = rbacBindings

	rdc := RBACDefinitionController{}
	rdc.kubernetesClientSet = client

	rdc.reconcileRbacDef(&rbacDef)

	clusterRoleBindings2, err := client.RbacV1().ClusterRoleBindings().List(listOptions)

	if err != nil {
		t.Fatal(err)
	}

	if len(clusterRoleBindings2.Items) != 1 {
		t.Fatal("Expected 1 cluster role binding after reconcile")
	}

	crb := clusterRoleBindings2.Items[0]

	if len(crb.Subjects) != 1 {
		t.Fatal("Expected 1 subject in cluster role binding")
	}

	subject := crb.Subjects[0]
	if subject.Kind != rbacv1.UserKind {
		t.Fatal("Expected subject to be user")
	}

	if subject.Name != "bar" {
		t.Fatal("Expected subject name to be bar")
	}

	roleRef := crb.RoleRef
	if roleRef.Kind != "ClusterRole" {
		t.Fatal("Expected roleRef kind to be ClusterRole")
	}

	if roleRef.Name != "admin" {
		t.Fatal("Expected roleRef name to be admin")
	}
}
