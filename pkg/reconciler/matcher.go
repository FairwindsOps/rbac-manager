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
	"reflect"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func crbMatches(existingCRB *rbacv1.ClusterRoleBinding, requestedCRB *rbacv1.ClusterRoleBinding) bool {
	if !metaMatches(&existingCRB.ObjectMeta, &requestedCRB.ObjectMeta) {
		return false
	}
	if !subjectsMatch(&existingCRB.Subjects, &requestedCRB.Subjects) {
		return false
	}

	if !roleRefMatches(&existingCRB.RoleRef, &requestedCRB.RoleRef) {
		return false
	}

	return true
}

func rbMatches(existingRB *rbacv1.RoleBinding, requestedRB *rbacv1.RoleBinding) bool {
	if !metaMatches(&existingRB.ObjectMeta, &requestedRB.ObjectMeta) {
		return false
	}
	if !subjectsMatch(&existingRB.Subjects, &requestedRB.Subjects) {
		return false
	}

	if !roleRefMatches(&existingRB.RoleRef, &requestedRB.RoleRef) {
		return false
	}

	return true
}

func saMatches(existingSA *v1.ServiceAccount, requestedSA *v1.ServiceAccount) bool {
	if metaMatches(&existingSA.ObjectMeta, &requestedSA.ObjectMeta) {
		if len(requestedSA.ImagePullSecrets) < 1 && existingSA.ImagePullSecrets == nil {
			return true
		}
		return reflect.DeepEqual(&existingSA.ImagePullSecrets, &requestedSA.ImagePullSecrets)
	}
	return false
}

func metaMatches(existingMeta *metav1.ObjectMeta, requestedMeta *metav1.ObjectMeta) bool {
	if existingMeta.Name != requestedMeta.Name {
		return false
	}

	if existingMeta.Namespace != requestedMeta.Namespace {
		return false
	}

	if !ownerRefsMatch(&existingMeta.OwnerReferences, &requestedMeta.OwnerReferences) {
		return false
	}

	return true
}

func ownerRefsMatch(existingOwnerRefs *[]metav1.OwnerReference, requestedOwnerRefs *[]metav1.OwnerReference) bool {
	requested := *requestedOwnerRefs
	existing := *existingOwnerRefs

	if len(requested) != len(existing) {
		return false
	}

	for index, existingOwnerRef := range existing {
		if !ownerRefMatches(&existingOwnerRef, &requested[index]) {
			return false
		}
	}

	return true
}

func ownerRefMatches(existingOwnerRef *metav1.OwnerReference, requestedOwnerRef *metav1.OwnerReference) bool {
	if existingOwnerRef.Kind != requestedOwnerRef.Kind {
		return false
	}

	if existingOwnerRef.Name != requestedOwnerRef.Name {
		return false
	}

	if existingOwnerRef.APIVersion != requestedOwnerRef.APIVersion {
		return false
	}

	return true
}

func subjectsMatch(existingSubjects *[]rbacv1.Subject, requestedSubjects *[]rbacv1.Subject) bool {
	rSubjects := *requestedSubjects
	eSubjects := *existingSubjects

	if len(eSubjects) != len(rSubjects) {
		return false
	}

	for index, existingSubject := range eSubjects {
		if !subjectMatches(&existingSubject, &rSubjects[index]) {
			return false
		}
	}

	return true
}

func subjectMatches(existingSubject *rbacv1.Subject, requestedSubject *rbacv1.Subject) bool {
	if existingSubject.Kind != requestedSubject.Kind {
		return false
	}

	if existingSubject.Name != requestedSubject.Name {
		return false
	}

	if existingSubject.Namespace != requestedSubject.Namespace {
		return false
	}

	return true
}

func roleRefMatches(existingRoleRef *rbacv1.RoleRef, requestedRoleRef *rbacv1.RoleRef) bool {
	if existingRoleRef.Kind != requestedRoleRef.Kind {
		return false
	}

	if existingRoleRef.Name != requestedRoleRef.Name {
		return false
	}

	return true
}
