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
	// TODO: Finda better place for this
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
			logrus.Errorf("Failed to process RbacDefinition: %v", err)
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

	ownerReferences := []metav1.OwnerReference{
		*metav1.NewControllerRef(rbacdef, schema.GroupVersionKind{
			Group:   v1beta1.SchemeGroupVersion.Group,
			Version: v1beta1.SchemeGroupVersion.Version,
			Kind:    "RbacDefinition",
		}),
	}

	logrus.Infof("RbacDefinition %v", rbacdef)

	if rbacdef.Spec.RbacBindings == nil {
		logrus.Warn("No RbacBindings defined")
		return nil
	}

	requestedClusterRoleBindings := []rbacv1.ClusterRoleBinding{}
	requestedRoleBindings := []rbacv1.RoleBinding{}

	for _, rbacBinding := range rbacdef.Spec.RbacBindings {
		if rbacBinding.ClusterRoleBindings != nil {
			for _, requestedCRB := range rbacBinding.ClusterRoleBindings {
				crbName := fmt.Sprintf("%v-%v-%v", rbacdef.Name, rbacBinding.Name, requestedCRB.ClusterRole)

				logrus.Infof("Processing CRB %v")

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

				if requestedRB.Namespace == "" {
					logrus.Error("Invalid role binding, namespace required")
					break
				}

				objectMeta.Namespace = requestedRB.Namespace

				if requestedRB.ClusterRole != "" {
					logrus.Infof("Processing Requested ClusterRole %v <> %v <> %v", requestedRB.ClusterRole, requestedRB.Namespace, requestedRB)
					requestedRoleName = requestedRB.ClusterRole
					roleRef = rbacv1.RoleRef{
						Kind: "ClusterRole",
						Name: requestedRB.ClusterRole,
					}
				} else if requestedRB.Role != "" {
					logrus.Infof("Processing Requested Role %v <> %v <> %v", requestedRB.Role, requestedRB.Namespace, requestedRB)
					requestedRoleName = fmt.Sprintf("%v-%v", requestedRB.Role, requestedRB.Namespace)
					roleRef = rbacv1.RoleRef{
						Kind: "Role",
						Name: requestedRB.Role,
					}
				} else {
					logrus.Error("Invalid role binding, role or clusterRole required")
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
			logrus.Infof("Attempting to create Cluster Role Binding: %v", requestedCRB)
			_, err := clientset.RbacV1().ClusterRoleBindings().Create(&requestedCRB)
			if err != nil {
				logrus.Errorf("Error creating Cluster Role Binding: %v", err)
			}
		} else {
			logrus.Infof("Cluster Role Binding already exists %v", requestedCRB)
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
				logrus.Infof("Attempting to delete Cluster Role Binding: %v", existingCRB)
				err := clientset.RbacV1().ClusterRoleBindings().Delete(existingCRB.Name, &metav1.DeleteOptions{})
				if err != nil {
					logrus.Errorf("Error deleting Cluster Role Binding: %v", err)
				}
			} else {
				logrus.Infof("Matches requested Cluster Role Binding: %v", err)
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
			logrus.Infof("Attempting to create Role Binding: %v", requestedRB)
			_, err := clientset.RbacV1().RoleBindings(requestedRB.ObjectMeta.Namespace).Create(&requestedRB)
			if err != nil {
				logrus.Errorf("Error creating Role Binding: %v", err)
			}
		} else {
			logrus.Infof("Role Binding already exists %v", requestedRB)
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
				logrus.Infof("Attempting to delete Role Binding %v", existingRB)
				err := clientset.RbacV1().ClusterRoleBindings().Delete(existingRB.Name, &metav1.DeleteOptions{})
				if err != nil {
					logrus.Infof("Error deleting Role Binding: %v", err)
				}
			} else {
				logrus.Infof("Matches requested Role Binding %v", existingRB)
			}
		}
	}

	return nil
}
