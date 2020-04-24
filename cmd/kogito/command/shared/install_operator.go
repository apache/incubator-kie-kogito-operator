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
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"

	"strings"

	"github.com/gobuffalo/packr/v2"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/version"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

const (
	defaultOperatorImageName = "quay.io/kiegroup/kogito-cloud-operator"
	boxDeployPath            = "../../../../deploy"
	fileOperatorYaml         = "operator.yaml"
	fileRoleYaml             = "role.yaml"
	fileRoleBindingYaml      = "role_binding.yaml"
	fileServiceAccountYaml   = "service_account.yaml"
	crdYAMLPattern           = "_crd.yaml"
	// skipOperatorInstallEnv if this is set to "true", we won't try to install the operator at all.
	// It's a flag indicating that we are running the operator locally with operator-sdk, thus not needed to install the Operator in the namespace
	skipOperatorInstallEnv = "SKIP_OPERATOR"
)

var (
	// DefaultOperatorImageNameTag is the default name of the Kogito operator image
	DefaultOperatorImageNameTag = fmt.Sprintf("%s:%s", defaultOperatorImageName, version.Version)
)

// SilentlyInstallOperatorIfNotExists attempts to install the operator and does not log a message if it is installed
func SilentlyInstallOperatorIfNotExists(namespace string, operatorImage string, client *client.Client) (installed bool, err error) {
	return InstallOperatorIfNotExists(namespace, operatorImage, client, false, false)
}

// InstallOperatorIfNotExists installs the operator using the deploy/*yaml and deploy/crds/*crds.yaml files, if the operator deployment is not in the given namespace.
// If the operator is available at the OperatorHub in OpenShift installations and not installed, tries to install the Operator via OLM Subscriptions.
// operatorImage can be an empty string. In this case, the empty string is the default value.
func InstallOperatorIfNotExists(namespace string, operatorImage string, cli *client.Client, warnIfInstalled bool, force bool) (installed bool, err error) {
	log := context.GetDefaultLogger()

	if util.GetBoolOSEnv(skipOperatorInstallEnv) {
		log.Infof("%s environment variable set to true, skipping operator installation process", skipOperatorInstallEnv)
		return true, nil
	}

	if len(operatorImage) == 0 {
		operatorImage = DefaultOperatorImageNameTag
	}

	if exists, err := infrastructure.CheckKogitoOperatorExists(cli, namespace); err != nil {
		return false, err
	} else if exists {
		if warnIfInstalled {
			log.Infof("Kogito Operator is already deployed in the namespace '%s', skipping ", namespace)
		}
		return true, nil
	}

	if available, err := isOperatorAvailableInOperatorHub(cli); err != nil {
		log.Warnf("Couldn't find if the Kogito Operator is available in the cluster: %s ", err)
	} else if available && !force {
		if err := installOperatorWithOperatorHub(namespace, cli); err != nil {
			log.Warnf("Couldn't install Kogito Operator via OperatorHub: %s ", err)
			return false, err
		}
		return true, nil
	}

	if force {
		log.Infof("Forcing installation of operator with custom image %s.", operatorImage)
	} else {
		log.Infof("Kogito Operator not found in the namespace '%s', trying to deploy it", namespace)
	}

	if err := installOperatorWithYamlFiles(operatorImage, namespace, cli); err != nil {
		return false, fmt.Errorf("Error while deploying Kogito Operator via template yaml files: %s ", err)
	}

	log.Infof("Kogito Operator successfully deployed in '%s' namespace", namespace)

	return true, nil
}

// installOperatorWithYamlFiles installs the Kogito Operator in clusters that doesn't have OperatorHub installed, such as OCP 3.x and vanilla Kubernetes
func installOperatorWithYamlFiles(image string, namespace string, cli *client.Client) error {
	box := packr.New("deploy", boxDeployPath)

	// creates all CRDs found in the deploy directory
	for _, crd := range getAllCRDsFileNames(box) {
		if err := decodeAndCreateKubeObject(box, crd, &apiextensionsv1beta1.CustomResourceDefinition{}, namespace, cli, nil); err != nil {
			return err
		}
	}

	if err := decodeAndCreateKubeObject(box, fileServiceAccountYaml, &v1.ServiceAccount{}, namespace, cli, nil); err != nil {
		return err
	}
	if err := decodeAndCreateKubeObject(box, fileRoleYaml, &rbac.Role{}, namespace, cli, nil); err != nil {
		return err
	}
	if err := decodeAndCreateKubeObject(box, fileRoleBindingYaml, &rbac.RoleBinding{}, namespace, cli, nil); err != nil {
		return err
	}
	if err := decodeAndCreateKubeObject(box, fileOperatorYaml, &apps.Deployment{}, namespace, cli, func(object interface{}) {
		if len(image) > 0 {
			object.(*apps.Deployment).Spec.Template.Spec.Containers[0].Image = image
		}
	}); err != nil {
		return err
	}

	return nil
}

func decodeAndCreateKubeObject(box *packr.Box, yamlDoc string, resourceRef meta.ResourceObject, namespace string, client *client.Client, beforeCreate func(object interface{})) error {
	dat, err := box.FindString(yamlDoc)
	if err != nil {
		return fmt.Errorf("Error reading file %s: %s ", yamlDoc, err)
	}

	if err = kubernetes.ResourceC(client).CreateFromYamlContent(dat, namespace, resourceRef, beforeCreate); err != nil {
		return fmt.Errorf("Error while creating resources from file '%s': %v ", yamlDoc, err)
	}

	return nil
}

// getAllCRDsFileNames reads all CRDs files from box
func getAllCRDsFileNames(box *packr.Box) []string {
	var crds []string
	for _, file := range box.List() {
		if strings.Contains(file, crdYAMLPattern) {
			crds = append(crds, file)
		}
	}
	return crds
}
