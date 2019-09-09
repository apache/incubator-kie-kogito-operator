package main

import (
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	kogitoOperatorInstallationInstruc = "https://github.com/kiegroup/kogito-cloud-operator/blob/master/README.md#installation"
)

func checkProjectExists(namespace string) error {
	log.Debugf("Checking if namespace '%s' exists", namespace)
	if ns, err := kubernetes.NamespaceC(kubeCli).Fetch(namespace); err != nil {
		return fmt.Errorf("Error while trying to fetch for the application context (namespace): %s", err)
	} else if ns == nil {
		return fmt.Errorf("Project %s not found. Try setting your project using 'kogito use-project NAME'", namespace)
	}
	log.Debugf("Namespace '%s' exists", namespace)
	return nil
}

func checkKogitoAppCRDExists() error {
	log.Debug("Checking if Kogito Operator CRD is installed")
	kogitocrd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: v1alpha1.KogitoAppCRDName,
		},
	}
	if exists, err := kubernetes.ResourceC(kubeCli).Fetch(kogitocrd); err != nil {
		return fmt.Errorf("Error while trying to look for the Kogito Operator: %s", err)
	} else if !exists {
		return fmt.Errorf("Couldn't find the Kogito Operator in your cluster, please follow the instructions in %s to install it", kogitoOperatorInstallationInstruc)
	}
	log.Debug("Kogito Operator CRD installed")
	return nil
}

// checkKogitoAppNotExists returns an error if the kogitoapp exists
func checkKogitoAppNotExists(name string, namespace string) error {
	log.Debugf("Checking if Kogito Service '%s' was deployed before on namespace %s", name, namespace)
	kogitoapp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if exists, err := kubernetes.ResourceC(kubeCli).Fetch(kogitoapp); err != nil {
		return fmt.Errorf("Error while trying to look for the KogitoApp: %s", err)
	} else if exists {
		return fmt.Errorf("Looks like a Kogito App with the name '%s' already exists in this context/namespace. Please try another name", name)
	}
	log.Debugf("No custom resource with name '%s' was found in the namespace '%s'", name, namespace)
	return nil
}

// checkKogitoAppExists will return an error if the Kogito Service DOES NOT exist in the project/namespace
func checkKogitoAppExists(name string, project string) error {
	log.Debugf("Checking if Kogito Service '%s' was deployed before on project %s", name, project)
	kogitoapp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: project,
		},
	}
	if exists, err := kubernetes.ResourceC(kubeCli).Fetch(kogitoapp); err != nil {
		return fmt.Errorf("Error while trying to look for the KogitoApp: %s", err)
	} else if !exists {
		return fmt.Errorf("Looks like a Kogito App with the name '%s' doesn't exist in this project. Please try another name", name)
	}
	log.Debugf("Custom resource with name '%s' was found in the project '%s'", name, project)
	return nil
}
