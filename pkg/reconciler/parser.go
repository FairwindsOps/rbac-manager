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
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	rbacmanagerv1beta1 "github.com/fairwindsops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
	"github.com/fairwindsops/rbac-manager/pkg/kube"
)

// Parser parses RBAC Definitions and determines the Kubernetes resources that it specifies
type Parser struct {
	Clientset                 kubernetes.Interface
	ownerRefs                 []metav1.OwnerReference
	parsedClusterRoleBindings []rbacv1.ClusterRoleBinding
	parsedRoleBindings        []rbacv1.RoleBinding
	parsedServiceAccounts     []v1.ServiceAccount
}

const ManagedPullSecretsAnnotationKey string = "rbacmanager.reactiveops.io/managed-pull-secrets"

// Parse determines the desired Kubernetes resources an RBAC Definition refers to
func (p *Parser) Parse(rbacDef rbacmanagerv1beta1.RBACDefinition) error {
	if rbacDef.RBACBindings == nil {
		slog.Warn("No RBACBindings defined")
		return nil
	}

	namespaces, err := p.Clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		slog.Debug("Error listing namespaces", "error", err)
		return err
	}

	for _, rbacBinding := range rbacDef.RBACBindings {
		namePrefix := rdNamePrefix(&rbacDef, &rbacBinding)
		err := p.parseRBACBinding(rbacBinding, namePrefix, namespaces)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) parseRBACBinding(rbacBinding rbacmanagerv1beta1.RBACBinding, namePrefix string, namespaces *v1.NamespaceList) error {
	for _, requestedSubject := range rbacBinding.Subjects {
		if requestedSubject.Kind == "ServiceAccount" {
			pullsecrets := []v1.LocalObjectReference{}
			for _, secret := range requestedSubject.ImagePullSecrets {
				pullsecrets = append(pullsecrets, v1.LocalObjectReference{Name: secret})
			}
			annotations := make(map[string]string)
			managedPullSecrets := strings.Join(requestedSubject.ImagePullSecrets, ",")
			annotations[ManagedPullSecretsAnnotationKey] = managedPullSecrets
			p.parsedServiceAccounts = append(p.parsedServiceAccounts, v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:            requestedSubject.Name,
					Namespace:       requestedSubject.Namespace,
					OwnerReferences: p.ownerRefs,
					Labels:          kube.Labels,
					Annotations:     annotations,
				},
				ImagePullSecrets:             pullsecrets,
				AutomountServiceAccountToken: requestedSubject.AutomountServiceAccountToken,
			})
		}
	}

	if rbacBinding.ClusterRoleBindings != nil {
		for _, requestedCRB := range rbacBinding.ClusterRoleBindings {
			err := p.parseClusterRoleBinding(requestedCRB, rbacBinding.Subjects, namePrefix)
			if err != nil {
				return err
			}
		}
	}

	if rbacBinding.RoleBindings != nil {
		for _, requestedRB := range rbacBinding.RoleBindings {
			err := p.parseRoleBinding(requestedRB, rbacBinding.Subjects, namePrefix, namespaces)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Parser) parseClusterRoleBinding(
	crb rbacmanagerv1beta1.ClusterRoleBinding, subjects []rbacmanagerv1beta1.Subject, prefix string) error {
	crbName := fmt.Sprintf("%v-%v", prefix, crb.ClusterRole)
	subs := managerSubjectsToRbacSubjects(subjects)

	p.parsedClusterRoleBindings = append(p.parsedClusterRoleBindings, rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            crbName,
			OwnerReferences: p.ownerRefs,
			Labels:          kube.Labels,
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: crb.ClusterRole,
		},
		Subjects: subs,
	})

	return nil
}

func (p *Parser) parseRoleBinding(
	rb rbacmanagerv1beta1.RoleBinding, subjects []rbacmanagerv1beta1.Subject, prefix string, namespaces *v1.NamespaceList) error {

	objectMeta := metav1.ObjectMeta{
		OwnerReferences: p.ownerRefs,
		Labels:          kube.Labels,
	}

	var requestedRoleName string
	var roleRef rbacv1.RoleRef

	if rb.ClusterRole != "" {
		slog.Debug("Processing Requested ClusterRole", "clusterRole", rb.ClusterRole, "namespace", rb.Namespace, "roleBinding", rb)
		requestedRoleName = rb.ClusterRole
		roleRef = rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: rb.ClusterRole,
		}
	} else if rb.Role != "" {
		slog.Debug("Processing Requested Role", "role", rb.Role, "namespace", rb.Namespace, "roleBinding", rb)
		requestedRoleName = fmt.Sprintf("%v-%v", rb.Role, rb.Namespace)
		roleRef = rbacv1.RoleRef{
			Kind: "Role",
			Name: rb.Role,
		}
	} else {
		return errors.New("invalid role binding, role or clusterRole required")
	}

	objectMeta.Name = fmt.Sprintf("%v-%v", prefix, requestedRoleName)

	if rb.NamespaceSelector.MatchLabels != nil || len(rb.NamespaceSelector.MatchExpressions) > 0 {
		slog.Debug("Processing Namespace Selector", "selector", rb.NamespaceSelector)

		selector, err := metav1.LabelSelectorAsSelector(&rb.NamespaceSelector)
		if err != nil {
			slog.Info("Error parsing label selector", "error", err)
			return err
		}

		for _, namespace := range namespaces.Items {
			// Lazy way to marshal map[] of labels in to a Set, which we can then match on.
			if selector.Matches(labels.Merge(namespace.Labels, namespace.Labels)) {
				slog.Debug("Adding Role Binding With Dynamic Namespace", "namespace", namespace.Name)

				om := objectMeta
				om.Namespace = namespace.Name
				subs := managerSubjectsToRbacSubjects(subjects)

				p.parsedRoleBindings = append(p.parsedRoleBindings, rbacv1.RoleBinding{
					ObjectMeta: om,
					RoleRef:    roleRef,
					Subjects:   subs,
				})
			}
		}

	} else if rb.Namespace != "" {
		objectMeta.Namespace = rb.Namespace
		subs := managerSubjectsToRbacSubjects(subjects)

		p.parsedRoleBindings = append(p.parsedRoleBindings, rbacv1.RoleBinding{
			ObjectMeta: objectMeta,
			RoleRef:    roleRef,
			Subjects:   subs,
		})

	} else {
		return errors.New("invalid role binding, namespace or namespace selector required")
	}

	return nil
}

func (p *Parser) hasNamespaceSelectors(rbacDef *rbacmanagerv1beta1.RBACDefinition) bool {
	for _, rbacBinding := range rbacDef.RBACBindings {
		for _, roleBinding := range rbacBinding.RoleBindings {
			if roleBinding.Namespace == "" {
				// Split these up instead of using || so we can test both paths.
				if roleBinding.NamespaceSelector.MatchLabels != nil {
					return true
				}
				if roleBinding.NamespaceSelector.MatchExpressions != nil {
					return true
				}
			}
		}
	}
	return false
}

func (p *Parser) parseClusterRoleBindings(rbacDef *rbacmanagerv1beta1.RBACDefinition) {
	for _, rbacBinding := range rbacDef.RBACBindings {
		for _, clusterRoleBinding := range rbacBinding.ClusterRoleBindings {
			namePrefix := rdNamePrefix(rbacDef, &rbacBinding)
			_ = p.parseClusterRoleBinding(clusterRoleBinding, rbacBinding.Subjects, namePrefix)
		}
	}
}

func (p *Parser) parseRoleBindings(rbacDef *rbacmanagerv1beta1.RBACDefinition, namespaces *v1.NamespaceList) {
	for _, rbacBinding := range rbacDef.RBACBindings {
		for _, roleBinding := range rbacBinding.RoleBindings {
			namePrefix := rdNamePrefix(rbacDef, &rbacBinding)
			_ = p.parseRoleBinding(roleBinding, rbacBinding.Subjects, namePrefix, namespaces)
		}
	}
}

func rdNamePrefix(rbacDef *rbacmanagerv1beta1.RBACDefinition, rbacBinding *rbacmanagerv1beta1.RBACBinding) string {
	return fmt.Sprintf("%v-%v", rbacDef.Name, rbacBinding.Name)
}

func managerSubjectsToRbacSubjects(subjects []rbacmanagerv1beta1.Subject) []rbacv1.Subject {
	var subs []rbacv1.Subject
	for _, sub := range subjects {
		subs = append(subs, rbacv1.Subject{
			Kind:      sub.Kind,
			APIGroup:  sub.APIGroup,
			Name:      sub.Name,
			Namespace: sub.Namespace,
		})
	}
	return subs
}
