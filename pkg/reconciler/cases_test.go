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
	rbacmanagerv1beta1 "github.com/fairwindsops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var trueVal bool = true
var falseVal bool = false

// Parse Subject into ServiceAccount
var saTestCases = []struct {
	name     string
	given    []rbacmanagerv1beta1.Subject
	expected []v1.ServiceAccount
}{
	{
		"AutomountServiceAccountToken is true",
		[]rbacmanagerv1beta1.Subject{{
			Subject:                      rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "robot", Namespace: "default"},
			AutomountServiceAccountToken: &trueVal,
		}},
		[]v1.ServiceAccount{{
			ObjectMeta:                   metav1.ObjectMeta{Name: "robot", Namespace: "default"},
			AutomountServiceAccountToken: &trueVal,
		}},
	},
	{
		"AutomountServiceAccountToken is false",
		[]rbacmanagerv1beta1.Subject{{
			Subject:                      rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "robot", Namespace: "default"},
			AutomountServiceAccountToken: &falseVal,
		}},
		[]v1.ServiceAccount{{
			ObjectMeta:                   metav1.ObjectMeta{Name: "robot", Namespace: "default"},
			AutomountServiceAccountToken: &falseVal,
		}},
	},
	{
		"AutomountServiceAccountToken is empty",
		[]rbacmanagerv1beta1.Subject{{
			Subject: rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "robot", Namespace: "default"},
		}},
		[]v1.ServiceAccount{{
			ObjectMeta:                   metav1.ObjectMeta{Name: "robot", Namespace: "default"},
			AutomountServiceAccountToken: nil,
		}},
	},
	{
		"ImagePullSecrets is empty",
		[]rbacmanagerv1beta1.Subject{{
			Subject:          rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "robot", Namespace: "default"},
			ImagePullSecrets: []string{},
		}},
		[]v1.ServiceAccount{{
			ObjectMeta:       metav1.ObjectMeta{Name: "robot", Namespace: "default"},
			ImagePullSecrets: nil,
		}},
	},
	{
		"ImagePullSecrets is not empty",
		[]rbacmanagerv1beta1.Subject{{
			Subject:          rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "robot", Namespace: "default"},
			ImagePullSecrets: []string{"secret-z", "secret-a"},
		}},
		[]v1.ServiceAccount{{
			ObjectMeta:       metav1.ObjectMeta{Name: "robot", Namespace: "default", Annotations: map[string]string{ManagedPullSecretsAnnotationKey: "secret-z,secret-a"}},
			ImagePullSecrets: []v1.LocalObjectReference{{Name: "secret-a"}, {Name: "secret-z"}},
		}},
	},
	{
		"Parsing multiple Service Accounts",
		[]rbacmanagerv1beta1.Subject{
			{Subject: rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "robot-a", Namespace: "default"}},
			{Subject: rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "robot-a", Namespace: "non-default"}},
			{Subject: rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "robot-z", Namespace: "default"}},
		},
		[]v1.ServiceAccount{
			{ObjectMeta: metav1.ObjectMeta{Name: "robot-z", Namespace: "default"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "robot-a", Namespace: "non-default"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "robot-a", Namespace: "default"}},
		},
	},
	{
		"Annotations are passed",
		[]rbacmanagerv1beta1.Subject{{
			Subject:  rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "robot", Namespace: "default"},
			Metadata: &metav1.ObjectMeta{Annotations: map[string]string{"annotation-a": "annotation-value-a", "annotation-b": "annotation-value-b"}},
		}},
		[]v1.ServiceAccount{{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "robot",
				Namespace:   "default",
				Annotations: map[string]string{"annotation-a": "annotation-value-a", "annotation-b": "annotation-value-b"},
			},
		}},
	},
	{
		"RBAC manager annotations may not be set directly",
		[]rbacmanagerv1beta1.Subject{{
			Subject:  rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "robot", Namespace: "default"},
			Metadata: &metav1.ObjectMeta{Annotations: map[string]string{ManagedPullSecretsAnnotationKey: "some-explicit-value", "annotation-a": "annotation-value-a", "annotation-b": "annotation-value-b"}},
		}},
		[]v1.ServiceAccount{{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "robot",
				Namespace:   "default",
				Annotations: map[string]string{"annotation-a": "annotation-value-a", "annotation-b": "annotation-value-b"},
			},
		}},
	},
	{
		"RBAC manager annotations may not be changed by overwriting it",
		[]rbacmanagerv1beta1.Subject{{
			Subject:          rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "robot", Namespace: "default"},
			Metadata:         &metav1.ObjectMeta{Annotations: map[string]string{ManagedPullSecretsAnnotationKey: "some-explicit-value"}},
			ImagePullSecrets: []string{"secret-z", "secret-a"},
		}},
		[]v1.ServiceAccount{{
			ObjectMeta:       metav1.ObjectMeta{Name: "robot", Namespace: "default", Annotations: map[string]string{ManagedPullSecretsAnnotationKey: "secret-z,secret-a"}},
			ImagePullSecrets: []v1.LocalObjectReference{{Name: "secret-a"}, {Name: "secret-z"}},
		}},
	},
	{
		"Labels are passed",
		[]rbacmanagerv1beta1.Subject{{
			Subject:  rbacv1.Subject{Kind: rbacv1.ServiceAccountKind, Name: "robot", Namespace: "default"},
			Metadata: &metav1.ObjectMeta{Labels: map[string]string{"label-a": "label-value-a", "label-b": "label-value-b"}},
		}},
		[]v1.ServiceAccount{{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "robot",
				Namespace: "default",
				Labels:    map[string]string{"label-a": "label-value-a", "label-b": "label-value-b"},
			},
		}},
	},
}
