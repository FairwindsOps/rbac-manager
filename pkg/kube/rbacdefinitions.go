/*
Copyright 2018 FairwindsOps Inc.

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
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	rbacmanagerv1beta1 "github.com/fairwindsops/rbac-manager/pkg/apis/rbacmanager/v1beta1"
)

// GetRbacDefinition returns an RbacDefinition for a specified name or an error
func GetRbacDefinition(name string) (rbacmanagerv1beta1.RBACDefinition, error) {
	rbacDef := rbacmanagerv1beta1.RBACDefinition{}

	client, err := getRbacDefClient()
	if err != nil {
		return rbacDef, err
	}

	err = client.Get().Resource("rbacdefinitions").Name(name).Do().Into(&rbacDef)

	return rbacDef, err
}

// GetRbacDefinitions returns an RbacDefinitionList or an error
func GetRbacDefinitions() (rbacmanagerv1beta1.RBACDefinitionList, error) {
	list := rbacmanagerv1beta1.RBACDefinitionList{}

	client, err := getRbacDefClient()
	if err != nil {
		return list, err
	}

	err = client.Get().Resource("rbacdefinitions").Do().Into(&list)

	return list, err
}

func getRbacDefClient() (*rest.RESTClient, error) {
	_ = rbacmanagerv1beta1.AddToScheme(scheme.Scheme)
	clientConfig := config.GetConfigOrDie()
	clientConfig.ContentConfig.GroupVersion = &rbacmanagerv1beta1.SchemeGroupVersion
	clientConfig.APIPath = "/apis"
	clientConfig.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}
	clientConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	return rest.UnversionedRESTClientFor(clientConfig)
}
