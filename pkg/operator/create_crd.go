package operator

import (
	"fmt"
	"time"

	rbacmanagerv1beta1 "github.com/reactiveops/rbac-manager/pkg/apis/rbac-manager/v1beta1"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// Create new RbacDefinition CRD
func createRbacDefinitionCRD() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return err
	}

	crd, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(rbacmanagerv1beta1.CRDName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			crdObject := &apiextensionsv1beta1.CustomResourceDefinition{
				ObjectMeta: metav1.ObjectMeta{
					Name: rbacmanagerv1beta1.CRDName,
				},
				Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
					Group:   rbacmanagerv1beta1.GroupName,
					Version: rbacmanagerv1beta1.Version,
					Scope:   apiextensionsv1beta1.NamespaceScoped,
					Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
						Plural: rbacmanagerv1beta1.ResourcePlural,
						Kind:   rbacmanagerv1beta1.ResourceKind,
					},
				},
			}

			_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crdObject)
			if err != nil {
				return err
			}
			logrus.Info("Created RbacDefinition CRD, waiting for it to be established")

			err = wait.Poll(500*time.Millisecond, 60*time.Second, func() (bool, error) {
				createdCRD, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(rbacmanagerv1beta1.CRDName, metav1.GetOptions{})
				if err != nil {
					return false, err
				}
				for _, cond := range createdCRD.Status.Conditions {
					switch cond.Type {
					case apiextensionsv1beta1.Established:
						if cond.Status == apiextensionsv1beta1.ConditionTrue {
							return true, nil
						}
					case apiextensionsv1beta1.NamesAccepted:
						if cond.Status == apiextensionsv1beta1.ConditionFalse {
							return false, fmt.Errorf("Name conflict: %v", cond.Reason)
						}
					}
				}
				return false, nil
			})

			if err != nil {
				deleteErr := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(rbacmanagerv1beta1.CRDName, nil)
				if deleteErr != nil {
					return errors.NewAggregate([]error{err, deleteErr})
				}
				return err
			}
		} else {
			return err
		}
	} else {
		logrus.Infof("RbacDefinition CRD already exists %v\n", crd.ObjectMeta.Name)
	}

	return nil
}
