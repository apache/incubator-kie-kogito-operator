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
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-cloud-operator/core/client"
	"github.com/kiegroup/kogito-cloud-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/core/manager"
	"github.com/kiegroup/kogito-cloud-operator/core/operator"
	"github.com/kiegroup/kogito-cloud-operator/internal"
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceCheckService is interface to check K8 resource existence
type ResourceCheckService interface {
	EnsureProject(kubeCli *client.Client, project string) (string, error)
	CheckKogitoRuntimeExists(kubeCli *client.Client, name string, namespace string) error
	CheckKogitoRuntimeNotExists(kubeCli *client.Client, name string, namespace string) error
	CheckKogitoBuildExists(kubeCli *client.Client, name string, project string) error
	CheckKogitoBuildNotExists(kubeCli *client.Client, name string, namespace string) error
	CheckKogitoInfraExists(kubeCli *client.Client, name string, namespace string) error
}

type resourceCheckServiceImpl struct{}

// NewResourceCheckService create and return resourceCheckServiceImpl value
func NewResourceCheckService() ResourceCheckService {
	return resourceCheckServiceImpl{}
}

// EnsureProject verifies whether the given project is a valid string in the context and whether it exists in the cluster
func (r resourceCheckServiceImpl) EnsureProject(kubeCli *client.Client, project string) (string, error) {
	return EnsureProject(kubeCli, project)
}

// EnsureProject verifies whether the given project is a valid string in the context and whether it exists in the cluster
func EnsureProject(kubeCli *client.Client, project string) (string, error) {
	log := context.GetDefaultLogger()
	projectInContext := GetCurrentNamespaceFromKubeConfig()
	// we don't care if it's not in the context if a project is given to us
	if len(projectInContext) == 0 && len(project) == 0 {
		return "", fmt.Errorf(message.ProjectNoContext)
	}
	if len(project) == 0 {
		project = projectInContext
	}
	if err := checkProjectExists(kubeCli, project); err != nil {
		return project, err
	}
	if project != projectInContext {
		if err := SetCurrentNamespaceToKubeConfig(project); err != nil {
			return "", err
		}
	}
	log.Debugf(message.ProjectUsingProject, project)
	return project, nil
}

// checkProjectExists ...
func checkProjectExists(kubeCli *client.Client, namespace string) error {
	log := context.GetDefaultLogger()
	log.Debugf("Checking if namespace '%s' exists", namespace)
	if ns, err := kubernetes.NamespaceC(kubeCli).Fetch(namespace); err != nil {
		return fmt.Errorf("Error while trying to fetch for the application context (namespace): %s ", err)
	} else if ns == nil {
		return fmt.Errorf("Project %s not found. Try setting your project using 'kogito use-project NAME' ", namespace)
	}
	log.Debugf("Namespace '%s' exists", namespace)
	return nil
}

// CheckKogitoRuntimeExists returns an error if the KogitoRuntime not exists
func (r resourceCheckServiceImpl) CheckKogitoRuntimeExists(kubeCli *client.Client, name string, namespace string) error {
	log := context.GetDefaultLogger()
	if exists, err := isKogitoRuntimeExists(kubeCli, name, namespace); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("Looks like a Kogito runtime with the name '%s' doesn't exist in this project. Please try another name ", name)
	} else {
		log.Debugf("Kogito runtime with name '%s' was found in the project '%s' ", name, namespace)
		return nil
	}
}

// CheckKogitoRuntimeNotExists returns an error if the KogitoRuntime exists
func (r resourceCheckServiceImpl) CheckKogitoRuntimeNotExists(kubeCli *client.Client, name string, namespace string) error {
	log := context.GetDefaultLogger()
	if exists, err := isKogitoRuntimeExists(kubeCli, name, namespace); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("Looks like a Kogito Runtime with the name '%s' already exists in this context/namespace. Please try another name ", name)
	} else {
		log.Debugf("Custom resource with name '%s' was not found in the project '%s' ", name, namespace)
		return nil
	}
}

func isKogitoRuntimeExists(kubeCli *client.Client, name string, namespace string) (bool, error) {
	log := context.GetDefaultLogger()
	log.Debugf("Checking if Kogito Service '%s' was deployed before on namespace %s", name, namespace)
	kogitoRuntime := &v1beta1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	exists, err := kubernetes.ResourceC(kubeCli).Fetch(kogitoRuntime)
	if err != nil {
		return false, fmt.Errorf("Error while trying to look for the KogitoRuntime: %s ", err)
	}
	return exists, nil
}

// CheckKogitoBuildExists returns an error if the KogitoBuild not exists
func (r resourceCheckServiceImpl) CheckKogitoBuildExists(kubeCli *client.Client, name string, namespace string) error {
	log := context.GetDefaultLogger()
	if exists, err := isKogitoBuildExists(kubeCli, name, namespace); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("Looks like a Kogito Build with the name '%s' doesn't exist in this project. Please try another name ", name)
	} else {
		log.Debugf("Kogito Build with name '%s' was found in the project '%s' ", name, namespace)
		return nil
	}
}

// CheckKogitoBuildNotExists returns an error if the KogitoBuild exists
func (r resourceCheckServiceImpl) CheckKogitoBuildNotExists(kubeCli *client.Client, name string, namespace string) error {
	log := context.GetDefaultLogger()
	if exists, err := isKogitoBuildExists(kubeCli, name, namespace); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("Looks like a Kogito Build with the name '%s' already exists in this context/namespace. Please try another name ", name)
	} else {
		log.Debugf("Kogito Build with name '%s' was not found in the project '%s' ", name, namespace)
		return nil
	}
}

func isKogitoBuildExists(kubeCli *client.Client, name string, namespace string) (bool, error) {
	log := context.GetDefaultLogger()
	log.Debugf("Checking if Kogito Build '%s' was deployed before on namespace %s", name, namespace)
	kogitoBuild := &v1beta1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	exists, err := kubernetes.ResourceC(kubeCli).Fetch(kogitoBuild)
	if err != nil {
		return false, fmt.Errorf("Error while trying to look for the KogitoBuild: %s ", err)
	}
	return exists, nil
}

// CheckKogitoInfraExists returns an error if the KogitoInfra not exists
func (r resourceCheckServiceImpl) CheckKogitoInfraExists(kubeCli *client.Client, name string, namespace string) error {
	coreLogger := logger.GetLogger("cli_kogito_infra")
	coreLogger = coreLogger.WithValues("name", name, "namespace", namespace)

	context := &operator.Context{
		Client: kubeCli,
		Log:    coreLogger,
		Scheme: internal.GetRegisteredSchema(),
	}
	infraHandler := internal.NewKogitoInfraHandler(context)
	infraManager := manager.NewKogitoInfraManager(context, infraHandler)
	_, err := infraManager.MustFetchKogitoInfraInstance(types.NamespacedName{Name: name, Namespace: namespace})
	return err
}
