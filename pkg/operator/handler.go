package operator

import (
	"fmt"
	"reflect"

	"github.com/reactiveops/rbac-manager/pkg/apis/rbac-manager/v1beta1"

	"github.com/operator-framework/operator-sdk/pkg/sdk/handler"
	"github.com/operator-framework/operator-sdk/pkg/sdk/types"
	"github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func NewHandler() handler.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx types.Context, event types.Event) error {
	switch o := event.Object.(type) {
	case *v1beta1.RbacDefinition:
		err := processRbacDefinition(o)
		if err != nil && !errors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create busybox pod : %v", err)
			return err
		}
	}
	return nil
}

func processRbacDefinition(rbacdef *v1beta1.RbacDefinition) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	listOptions := metav1.ListOptions{LabelSelector: "rbac-manager=reactiveops"}
	labels := map[string]string{
		"rbac-manager": "reactiveops",
	}

	existingManagedClusterRoleBindings, err := clientset.RbacV1().ClusterRoleBindings().List(listOptions)
	existingManagedRoleBindings, err := clientset.RbacV1().RoleBindings("").List(listOptions)

	fmt.Printf("existingManagedRoleBindings ====> %v", existingManagedRoleBindings)

	ownerReferences := []metav1.OwnerReference{
		*metav1.NewControllerRef(rbacdef, schema.GroupVersionKind{
			Group:   v1beta1.SchemeGroupVersion.Group,
			Version: v1beta1.SchemeGroupVersion.Version,
			Kind:    "RbacDefinition",
		}),
	}

	fmt.Println("============ 1 ==============")
	fmt.Printf("%v", rbacdef)
	fmt.Println("============ 2 ==============")

	if rbacdef.Spec.RbacBindings == nil {
		fmt.Println("No RbacBindings defined")
		return nil
	}

	requestedClusterRoleBindings := []rbacv1.ClusterRoleBinding{}
	requestedRoleBindings := []rbacv1.RoleBinding{}

	for _, rbacBinding := range rbacdef.Spec.RbacBindings {
		if rbacBinding.ClusterRoleBindings != nil {
			for _, requestedCRB := range rbacBinding.ClusterRoleBindings {
				crbName := fmt.Sprintf("%v-%v-%v", rbacdef.Name, rbacBinding.Name, requestedCRB.ClusterRole)

				fmt.Printf("=> crbName ====> %v\n", crbName)

				requestedClusterRoleBindings = append(requestedClusterRoleBindings, rbacv1.ClusterRoleBinding{
					ObjectMeta: metav1.ObjectMeta{
						Name:            crbName,
						OwnerReferences: ownerReferences,
						Labels:          labels,
					},
					RoleRef: rbacv1.RoleRef{
						Kind: "ClusterRole",
						Name: requestedCRB.ClusterRole,
					},
					Subjects: rbacBinding.Subjects,
				})
			}
		}

		if rbacBinding.RoleBindings != nil {
			for _, requestedRB := range rbacBinding.RoleBindings {
				objectMeta := metav1.ObjectMeta{
					OwnerReferences: ownerReferences,
					Labels:          labels,
				}

				var requestedRoleName string
				var roleRef rbacv1.RoleRef

				fmt.Printf("HELLO requested role ===================> %v <> %v <> %v", requestedRB.Role, requestedRB.Namespace, requestedRB)

				if requestedRB.Namespace == "" {
					fmt.Printf("Invalid role binding, namespace required\n")
					break
				}

				objectMeta.Namespace = requestedRB.Namespace

				if requestedRB.ClusterRole != "" {
					requestedRoleName = requestedRB.ClusterRole
					roleRef = rbacv1.RoleRef{
						Kind: "ClusterRole",
						Name: requestedRB.ClusterRole,
					}
				} else if requestedRB.Role != "" {
					fmt.Printf("requested role ===================> %v <> %v <> %v", requestedRB.Role, requestedRB.Namespace, requestedRB)
					requestedRoleName = fmt.Sprintf("%v-%v", requestedRB.Role, requestedRB.Namespace)
					roleRef = rbacv1.RoleRef{
						Kind: "Role",
						Name: requestedRB.Role,
					}
				} else {
					fmt.Printf("Invalid role binding, role or clusterRole required\n")
					break
				}

				objectMeta.Name = fmt.Sprintf("%v-%v-%v", rbacdef.Name, rbacBinding.Name, requestedRoleName)

				requestedRoleBindings = append(requestedRoleBindings, rbacv1.RoleBinding{
					ObjectMeta: objectMeta,
					RoleRef:    roleRef,
					Subjects:   rbacBinding.Subjects,
				})
			}
		}
	}

	matchingClusterRoleBindings := []rbacv1.ClusterRoleBinding{}

	for _, requestedCRB := range requestedClusterRoleBindings {
		alreadyExists := false
		for _, existingCRB := range existingManagedClusterRoleBindings.Items {
			if crbMatches(&existingCRB, &requestedCRB) {
				alreadyExists = true
				matchingClusterRoleBindings = append(matchingClusterRoleBindings, existingCRB)
				break
			}
		}

		if !alreadyExists {
			fmt.Printf("Attempting to create %v\n", requestedCRB)
			crb, err := clientset.RbacV1().ClusterRoleBindings().Create(&requestedCRB)
			fmt.Printf("err ===> %v\n", err)
			fmt.Printf("crb ===> %v\n", crb)
		} else {
			fmt.Printf("Already exists %v\n", requestedCRB)
		}
	}

	for _, existingCRB := range existingManagedClusterRoleBindings.Items {
		if reflect.DeepEqual(existingCRB.OwnerReferences, ownerReferences) {
			matchingRequest := false
			for _, requestedCRB := range matchingClusterRoleBindings {
				if crbMatches(&existingCRB, &requestedCRB) {
					matchingRequest = true
					break
				}
			}

			if !matchingRequest {
				fmt.Printf("Attempting to delete %v\n", existingCRB)
				err := clientset.RbacV1().ClusterRoleBindings().Delete(existingCRB.Name, &metav1.DeleteOptions{})
				fmt.Printf("err ===> %v\n", err)
			} else {
				fmt.Printf("Matching request found %v\n", existingCRB)
			}
		}
	}

	matchingRoleBindings := []rbacv1.RoleBinding{}

	for _, requestedRB := range requestedRoleBindings {
		alreadyExists := false
		for _, existingRB := range existingManagedRoleBindings.Items {
			if rbMatches(&existingRB, &requestedRB) {
				alreadyExists = true
				matchingRoleBindings = append(matchingRoleBindings, existingRB)
				break
			}
		}

		if !alreadyExists {
			fmt.Printf("Attempting to create %v\n", requestedRB)
			rb, err := clientset.RbacV1().RoleBindings(requestedRB.ObjectMeta.Namespace).Create(&requestedRB)
			fmt.Printf("err ===> %v\n", err)
			fmt.Printf("rb ===> %v\n", rb)
		} else {
			fmt.Printf("Already exists %v\n", requestedRB)
		}
	}

	for _, existingRB := range existingManagedRoleBindings.Items {
		if reflect.DeepEqual(existingRB.OwnerReferences, ownerReferences) {
			matchingRequest := false
			for _, requestedRB := range matchingRoleBindings {
				if rbMatches(&existingRB, &requestedRB) {
					matchingRequest = true
					break
				}
			}

			if !matchingRequest {
				fmt.Printf("Attempting to delete %v\n", existingRB)
				err := clientset.RbacV1().ClusterRoleBindings().Delete(existingRB.Name, &metav1.DeleteOptions{})
				fmt.Printf("err ===> %v\n", err)
			} else {
				fmt.Printf("Matching request found %v\n", existingRB)
			}
		}
	}

	return nil
}
