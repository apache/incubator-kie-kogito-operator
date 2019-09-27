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

package shared

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/operator"
	v1 "k8s.io/api/apps/v1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnsureProject will check if the given project is a valid string in the context and if exists in the cluster
func EnsureProject(kubeCli *client.Client, config context.Configuration, project string) (string, error) {
	var err error
	if project, err = CheckProjectLocally(config, project); err != nil {
		return project, err
	}
	if err = CheckProjectExists(kubeCli, project); err != nil {
		return project, err
	}
	log.Debugf("Using namespace %s", project)
	config.Namespace = project
	config.Save()
	return project, nil
}

// CheckProjectExists ...
func CheckProjectExists(kubeCli *client.Client, namespace string) error {
	log.Debugf("Checking if namespace '%s' exists", namespace)
	if ns, err := kubernetes.NamespaceC(kubeCli).Fetch(namespace); err != nil {
		return fmt.Errorf("Error while trying to fetch for the application context (namespace): %s ", err)
	} else if ns == nil {
		return fmt.Errorf("Project %s not found. Try setting your project using 'kogito use-project NAME' ", namespace)
	}
	log.Debugf("Namespace '%s' exists", namespace)
	return nil
}

// CheckKogitoOperatorExists ...
func CheckKogitoOperatorExists(kubeCli *client.Client, namespace string) (bool, error) {
	log.Debug("Checking if Kogito Operator is deployed")
	operatorDeployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operator.Name,
			Namespace: namespace,
		},
	}

	if exists, err := kubernetes.ResourceC(kubeCli).Fetch(operatorDeployment); err != nil {
		return false, fmt.Errorf("Error while trying to look for Kogito Operator installation: %s ", err)
	} else if !exists {
		return false, nil
	}

	if operatorDeployment.Status.AvailableReplicas == 0 {
		return true, fmt.Errorf("Kogito Operator seems to be created in the namespace '%s', but there's no available pods replicas deployed ", namespace)
	}

	return true, nil
}

// CheckKogitoAppNotExists returns an error if the kogitoapp exists
func CheckKogitoAppNotExists(kubeCli *client.Client, name string, namespace string) error {
	log.Debugf("Checking if Kogito Service '%s' was deployed before on namespace %s", name, namespace)
	kogitoapp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if exists, err := kubernetes.ResourceC(kubeCli).Fetch(kogitoapp); err != nil {
		return fmt.Errorf("Error while trying to look for the KogitoApp: %s ", err)
	} else if exists {
		return fmt.Errorf("Looks like a Kogito App with the name '%s' already exists in this context/namespace. Please try another name ", name)
	}
	log.Debugf("No custom resource with name '%s' was found in the namespace '%s'", name, namespace)
	return nil
}

// CheckKogitoAppExists will return an error if the Kogito Service DOES NOT exist in the project/namespace
func CheckKogitoAppExists(kubeCli *client.Client, name string, project string) error {
	log.Debugf("Checking if Kogito Service '%s' was deployed before on project %s", name, project)
	kogitoapp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: project,
		},
	}
	if exists, err := kubernetes.ResourceC(kubeCli).Fetch(kogitoapp); err != nil {
		return fmt.Errorf("Error while trying to look for the KogitoApp: %s ", err)
	} else if !exists {
		return fmt.Errorf("Looks like a Kogito App with the name '%s' doesn't exist in this project. Please try another name ", name)
	}
	log.Debugf("Custom resource with name '%s' was found in the project '%s' ", name, project)
	return nil
}
