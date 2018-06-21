package v1beta1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/reactiveops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
	. "github.com/reactiveops/rbac-manager/pkg/client/clientset/versioned/typed/rbacmanager/v1beta1"
)

// EDIT THIS FILE!
// Created by "kubebuilder create resource" for you to implement the RBACDefinition resource tests

var _ = Describe("RBACDefinition", func() {
	var instance RBACDefinition
	var expected RBACDefinition
	var client RBACDefinitionInterface

	BeforeEach(func() {
		instance = RBACDefinition{}
		instance.Name = "instance-1"
		instance.RBACBindings = []RBACBinding{}

		expected = instance
	})

	AfterEach(func() {
		client.Delete(instance.Name, &metav1.DeleteOptions{})
	})

	// INSERT YOUR CODE HERE - add more "Describe" tests

	// Automatically created storage tests
	Describe("when sending a storage request", func() {
		Context("for a valid config", func() {
			It("should provide CRUD access to the object", func() {
				client = cs.RbacmanagerV1beta1().RBACDefinitions()

				By("returning success from the create request")
				actual, err := client.Create(&instance)
				Expect(err).ShouldNot(HaveOccurred())

				By("defaulting the expected fields")
				Expect(actual.RBACBindings).To(Equal(expected.RBACBindings))

				By("returning the item for list requests")
				result, err := client.List(metav1.ListOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(result.Items).To(HaveLen(1))
				Expect(result.Items[0].RBACBindings).To(Equal(expected.RBACBindings))

				By("returning the item for get requests")
				actual, err = client.Get(instance.Name, metav1.GetOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(actual.RBACBindings).To(Equal(expected.RBACBindings))

				By("deleting the item for delete requests")
				err = client.Delete(instance.Name, &metav1.DeleteOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				result, err = client.List(metav1.ListOptions{})
				Expect(err).ShouldNot(HaveOccurred())
				Expect(result.Items).To(HaveLen(0))
			})
		})
	})
})
