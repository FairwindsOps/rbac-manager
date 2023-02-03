/*
Copyright 2019 FairwindsOps Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kube

import (
	"os"
	"regexp"
	"strconv"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// LabelKey is the key of the key/value pair given to all resources managed by RBAC Manager
const LabelKey = "rbac-manager"

// LabelValue is the value of the key/value pair given to all resources managed by RBAC Manager
const LabelValue = "reactiveops"

// Labels is the key/value pair given to all resources managed by RBAC Manager
var Labels = map[string]string{LabelKey: LabelValue}

// ListOptions is the default set of options to find resources managed by RBAC Manager
var ListOptions = metav1.ListOptions{LabelSelector: LabelKey + "=" + LabelValue}

// GetClientsetOrDie returns a new Kubernetes Clientset or dies
func GetClientsetOrDie() *kubernetes.Clientset {
	kubeConf, err := config.GetConfig()

	if err != nil {
		logrus.Error(err, "unable to get Kubernetes client config")
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(kubeConf)

	if err != nil {
		logrus.Error(err, "unable to get Kubernetes clientset")
		os.Exit(1)
	}

	return clientset
}

// GetKubeVersion returns the major and minor version of the Kubernetes cluster
func GetKubeVersion() (major, minor int) {
	clientset := GetClientsetOrDie()
	version, err := clientset.ServerVersion()
	if err != nil {
		logrus.Error(err, "unable to get Kubernetes version")
		os.Exit(1)
	}

	reVersion := regexp.MustCompile(`^\d+`)
	version.Major = reVersion.FindString(version.Major)
	version.Minor = reVersion.FindString(version.Minor)

	majorInt, err := strconv.Atoi(version.Major)
	if err != nil {
		logrus.Error(err, "unable to convert server major version to int")
		os.Exit(1)
	}
	minorInt, err := strconv.Atoi(version.Minor)
	if err != nil {
		logrus.Error(err, "unable to convert server minor version to int")
		os.Exit(1)
	}
	return majorInt, minorInt
}
