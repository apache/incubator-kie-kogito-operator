// Copyright 2019 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ensureProject will check if the given project is a valid string in the context and if exists in the cluster
func ensureProject(project string) (string, error) {
	if project, err := checkProjectLocally(project); err != nil {
		return project, err
	}
	if err := checkProjectExists(project); err != nil {
		return project, err
	}
	log.Debugf("Using namespace %s", project)
	return project, nil
}

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

func checkKogitoCRDExists(crd *apiextensionsv1beta1.CustomResourceDefinition) error {
	log.Debugf("Checking if Kogito %s CRD is installed", crd.Name)
	if exists, err := kubernetes.ResourceC(kubeCli).Fetch(crd); err != nil {
		return fmt.Errorf("Error while trying to look for the Kogito CRD: %s", err)
	} else if !exists {
		return fmt.Errorf("Couldn't find the Kogito CRD %s in your cluster, please follow the instructions in https://github.com/kiegroup/kogito-cloud-operator#installation to install it", crd.Name)
	}
	log.Debugf("Kogito CRD %s installed", crd.Name)
	return nil
}

func checkKogitoDataIndexCRDExists() error {
	return checkKogitoCRDExists(&apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: v1alpha1.KogitoDataIndexCRDName,
		},
	})
}

func checkKogitoAppCRDExists() error {
	return checkKogitoCRDExists(&apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: v1alpha1.KogitoAppCRDName,
		},
	})
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

func checkKogitoDataIndexNotExists(namespace string) error {
	dataIndexList := v1alpha1.KogitoDataIndexList{}
	if err := kubernetes.ResourceC(kubeCli).ListWithNamespace(namespace, &dataIndexList); err != nil {
		return fmt.Errorf("Error while trying to look for the KogitoDataIndex: %s", err)
	}
	if len(dataIndexList.Items) > 0 {
		return fmt.Errorf("The namespace %s already has a Kogito Data Index service. No more than one data index service is allowed per namespace", namespace)
	}
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
