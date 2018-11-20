package rbacdefinition

import (
	"github.com/stretchr/testify/assert"
	"testing"

	rbacmanagerv1beta1 "github.com/reactiveops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestParseEmpty(t *testing.T) {
	client := fake.NewSimpleClientset()
	rbacDef := rbacmanagerv1beta1.RBACDefinition{}
	rbacDef.Name = "empty-example"
	testEmpty(t, client, rbacDef.Name)
}

func newParseTest(t *testing.T, client *fake.Clientset, rbacDef rbacmanagerv1beta1.RBACDefinition, expectedRb []rbacv1.RoleBinding, expectedCrb []rbacv1.ClusterRoleBinding, expectedSa []corev1.ServiceAccount) {
	rdc := RBACDefinitionController{}
	rdc.kubernetesClientSet = client

	rdp := rbacDefinitionParser{
		k8sClientSet: client,
		listOptions:  metav1.ListOptions{LabelSelector: "rbac-manager=reactiveops"},
	}
	rdp.parse(rbacDef)
	expectParsedRB(t, rdp, expectedRb)
	expectParsedCRB(t, rdp, expectedCrb)
	expectParsedSA(t, rdp, expectedSa)
}

func testEmpty(t *testing.T, client *fake.Clientset, name string) {
	rbacDef := rbacmanagerv1beta1.RBACDefinition{}
	rbacDef.Name = name
	rbacDef.RBACBindings = []rbacmanagerv1beta1.RBACBinding{}

	newParseTest(t, client, rbacDef, []rbacv1.RoleBinding{}, []rbacv1.ClusterRoleBinding{}, []corev1.ServiceAccount{})
}

func expectParsedCRB(t *testing.T, rdp rbacDefinitionParser, expected []rbacv1.ClusterRoleBinding) {
	assert.Len(t, rdp.parsedClusterRoleBindings, len(expected), "Expected length to match")

	for _, expectedCrb := range expected {
		matchFound := false
		for _, actualCrb := range rdp.parsedClusterRoleBindings {
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

func expectParsedRB(t *testing.T, rdp rbacDefinitionParser, expected []rbacv1.RoleBinding) {
	assert.Len(t, rdp.parsedRoleBindings, len(expected), "Expected length to match")

	for _, expectedRb := range expected {
		matchFound := false
		for _, actualRb := range rdp.parsedRoleBindings {
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

func expectParsedSA(t *testing.T, rdp rbacDefinitionParser, expected []corev1.ServiceAccount) {
	assert.Len(t, rdp.parsedServiceAccounts, len(expected), "Expected length to match")

	for _, expectedSa := range expected {
		matchFound := false
		for _, actualSa := range rdp.parsedServiceAccounts {
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
