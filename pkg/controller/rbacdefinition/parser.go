package rbacdefinition

import (
	"errors"
	"fmt"

	rbacmanagerv1beta1 "github.com/reactiveops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
	logrus "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type rbacDefinitionParser struct {
	rbacDef                   rbacmanagerv1beta1.RBACDefinition
	k8sClientSet              kubernetes.Interface
	listOptions               metav1.ListOptions
	labels                    map[string]string
	ownerRefs                 []metav1.OwnerReference
	parsedClusterRoleBindings []rbacv1.ClusterRoleBinding
	parsedRoleBindings        []rbacv1.RoleBinding
	parsedServiceAccounts     []v1.ServiceAccount
}

func (rdp *rbacDefinitionParser) parse() error {
	if rdp.rbacDef.RBACBindings == nil {
		logrus.Warn("No RBACBindings defined")
		return nil
	}

	for _, rbacBinding := range rdp.rbacDef.RBACBindings {
		err := rdp.parseRBACBinding(rbacBinding)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rdp *rbacDefinitionParser) parseRBACBinding(rbacBinding rbacmanagerv1beta1.RBACBinding) error {
	for _, requestedSubject := range rbacBinding.Subjects {
		if requestedSubject.Kind == "ServiceAccount" {
			rdp.parsedServiceAccounts = append(rdp.parsedServiceAccounts, v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:            requestedSubject.Name,
					Namespace:       requestedSubject.Namespace,
					OwnerReferences: rdp.ownerRefs,
					Labels:          rdp.labels,
				},
			})
		}

		namePrefix := fmt.Sprintf("%v-%v", rdp.rbacDef.Name, rbacBinding.Name)

		if rbacBinding.ClusterRoleBindings != nil {
			for _, requestedCRB := range rbacBinding.ClusterRoleBindings {
				err := rdp.parseClusterRoleBinding(requestedCRB, rbacBinding.Subjects, namePrefix)
				if err != nil {
					return err
				}
			}
		}

		if rbacBinding.RoleBindings != nil {
			for _, requestedRB := range rbacBinding.RoleBindings {
				err := rdp.parseRoleBinding(requestedRB, rbacBinding.Subjects, namePrefix)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (rdp *rbacDefinitionParser) parseClusterRoleBinding(
	crb rbacmanagerv1beta1.ClusterRoleBinding, subjects []rbacv1.Subject, prefix string) error {
	crbName := fmt.Sprintf("%v-%v", prefix, crb.ClusterRole)

	rdp.parsedClusterRoleBindings = append(rdp.parsedClusterRoleBindings, rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            crbName,
			OwnerReferences: rdp.ownerRefs,
			Labels:          rdp.labels,
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: crb.ClusterRole,
		},
		Subjects: subjects,
	})

	return nil
}

func (rdp *rbacDefinitionParser) parseRoleBinding(
	rb rbacmanagerv1beta1.RoleBinding, subjects []rbacv1.Subject, prefix string) error {

	objectMeta := metav1.ObjectMeta{
		OwnerReferences: rdp.ownerRefs,
		Labels:          rdp.labels,
	}

	var requestedRoleName string
	var roleRef rbacv1.RoleRef

	if rb.ClusterRole != "" {
		logrus.Debugf("Processing Requested ClusterRole %v <> %v <> %v", rb.ClusterRole, rb.Namespace, rb)
		requestedRoleName = rb.ClusterRole
		roleRef = rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: rb.ClusterRole,
		}
	} else if rb.Role != "" {
		logrus.Debugf("Processing Requested Role %v <> %v <> %v", rb.Role, rb.Namespace, rb)
		requestedRoleName = fmt.Sprintf("%v-%v", rb.Role, rb.Namespace)
		roleRef = rbacv1.RoleRef{
			Kind: "Role",
			Name: rb.Role,
		}
	} else {
		return errors.New("Invalid role binding, role or clusterRole required")
	}

	objectMeta.Name = fmt.Sprintf("%v-%v", prefix, requestedRoleName)

	if rb.NamespaceSelector != nil {
		listOptions := rdp.listOptions
		logrus.Debugf("Processing Namespace Selector %v", rb.NamespaceSelector)

		listOptions.LabelSelector = rb.NamespaceSelector.String()
		namespaces, err := rdp.k8sClientSet.CoreV1().Namespaces().List(listOptions)
		if err != nil {
			return err
		}

		for _, namespace := range namespaces.Items {
			logrus.Debugf("Adding Role Binding With Dynamic Namespace %v", namespace.Name)

			om := objectMeta
			om.Namespace = namespace.Name

			rdp.parsedRoleBindings = append(rdp.parsedRoleBindings, rbacv1.RoleBinding{
				ObjectMeta: om,
				RoleRef:    roleRef,
				Subjects:   subjects,
			})
		}

	} else if rb.Namespace != "" {
		objectMeta.Namespace = rb.Namespace

		rdp.parsedRoleBindings = append(rdp.parsedRoleBindings, rbacv1.RoleBinding{
			ObjectMeta: objectMeta,
			RoleRef:    roleRef,
			Subjects:   subjects,
		})

	} else {
		return errors.New("Invalid role binding, namespace or namespace selector required")
	}

	return nil
}
