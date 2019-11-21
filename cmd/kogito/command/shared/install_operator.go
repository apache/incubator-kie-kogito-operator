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

	"strings"

	"github.com/gobuffalo/packr/v2"

	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/operator"
	"github.com/kiegroup/kogito-cloud-operator/version"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	defaultOperatorImageName = "quay.io/kiegroup/kogito-cloud-operator"
	boxDeployPath            = "../../../../deploy"
	fileOperatorYaml         = "operator.yaml"
	fileRoleYaml             = "role.yaml"
	fileRoleBindingYaml      = "role_binding.yaml"
	fileServiceAccountYaml   = "service_account.yaml"
	crdYAMLpattern           = "_crd.yaml"
	docURLInstallOperator    = "https://github.com/kiegroup/kogito-cloud-operator#via-operatorhub-automatically"
)

var (
	// DefaultOperatorImageNameTag is the default name of the Kogito operator image
	DefaultOperatorImageNameTag = fmt.Sprintf("%s:%s", defaultOperatorImageName, version.Version)
)

// SilentlyInstallOperatorIfNotExists attempts to install the operator and does not log a message if it is installed
func SilentlyInstallOperatorIfNotExists(namespace string, operatorImage string, client *client.Client) (installed bool, err error) {
	return MustInstallOperatorIfNotExists(namespace, operatorImage, client, true)
}

// MustInstallOperatorIfNotExists installs the operator using the deploy/*yaml and deploy/crds/*crds.yaml files, if the operator deployment is not in the given namespace.
// If the operator is available at the OperatorHub in OpenShift installations and not installed, MustInstallOperatorIfNotExists generates a warning message and stops the flow
// operatorImage can be an empty string. In this case, the empty string is the default value.
func MustInstallOperatorIfNotExists(namespace string, operatorImage string, client *client.Client, silence bool) (installed bool, err error) {
	log := context.GetDefaultLogger()

	if len(operatorImage) == 0 {
		operatorImage = DefaultOperatorImageNameTag
	}

	if exists, err := checkKogitoOperatorExists(client, namespace); err != nil {
		return false, err
	} else if exists {
		if !silence {
			log.Infof("Kogito Operator is already deployed in the namespace '%s', skipping ", namespace)
		}
		return true, nil
	}

	if available, err := isOperatorAvailableInOperatorHub(client); err != nil {
		log.Warnf("Couldn't find if the Kogito Operator is available in the cluster: %s ", err)
	} else if available {
		log.Warnf("Couldn't install the Kogito Operator automatically. The operator is available in the OperatorHub, please install it manually via Web Console. See: %s ", docURLInstallOperator)
		return false, nil
	}

	log.Infof("Kogito Operator not found in the namespace '%s', trying to deploy it", namespace)

	if err := installOperatorWithYamlFiles(operatorImage, namespace, client); err != nil {
		return false, fmt.Errorf("Error while deploying Kogito Operator via template yaml files: %s ", err)
	}

	log.Infof("Kogito Operator successfully deployed in '%s' namespace", namespace)

	return true, nil
}

func installOperatorWithYamlFiles(image string, namespace string, client *client.Client) error {
	box := packr.New("deploy", boxDeployPath)

	// creates all CRDs found in the deploy directory
	for _, crd := range getAllCRDsFileNames(box) {
		if err := decodeAndCreateKubeObject(box, crd, &apiextensionsv1beta1.CustomResourceDefinition{}, namespace, client, nil); err != nil {
			return err
		}
	}

	if err := decodeAndCreateKubeObject(box, fileServiceAccountYaml, &v1.ServiceAccount{}, namespace, client, nil); err != nil {
		return err
	}
	if err := decodeAndCreateKubeObject(box, fileRoleYaml, &rbac.Role{}, namespace, client, nil); err != nil {
		return err
	}
	if err := decodeAndCreateKubeObject(box, fileRoleBindingYaml, &rbac.RoleBinding{}, namespace, client, nil); err != nil {
		return err
	}
	if err := decodeAndCreateKubeObject(box, fileOperatorYaml, &apps.Deployment{}, namespace, client, func(object interface{}) {
		if len(image) > 0 {
			object.(*apps.Deployment).Spec.Template.Spec.Containers[0].Image = image
		}
	}); err != nil {
		return err
	}

	return nil
}

func decodeAndCreateKubeObject(box *packr.Box, yamlDoc string, resourceRef meta.ResourceObject, namespace string, client *client.Client, beforeCreate func(object interface{})) error {
	log := context.GetDefaultLogger()
	dat, err := box.FindString(yamlDoc)
	if err != nil {
		return fmt.Errorf("Error reading file %s: %s ", yamlDoc, err)
	}

	docs := strings.Split(dat, "---")
	for _, doc := range docs {
		if err := yaml.NewYAMLOrJSONDecoder(strings.NewReader(doc), len([]byte(doc))).Decode(resourceRef); err != nil {
			return fmt.Errorf("Error while unmarshal file '%s': %s ", yamlDoc, err)
		}
		resourceRef.SetNamespace(namespace)
		resourceRef.SetResourceVersion("")
		resourceRef.SetLabels(map[string]string{"app": operator.Name})

		log.Debugf("Will create a new resource '%s' with name %s on %s ", resourceRef.GetObjectKind().GroupVersionKind().Kind, resourceRef.GetName(), resourceRef.GetNamespace())
		if beforeCreate != nil {
			beforeCreate(resourceRef)
		}
		if _, err := kubernetes.ResourceC(client).CreateIfNotExists(resourceRef); err != nil {
			return fmt.Errorf("Error creating object %s: %s ", resourceRef.GetName(), err)
		}
	}

	return nil
}

// getAllCRDsFileNames reads all CRDs files from box
func getAllCRDsFileNames(box *packr.Box) []string {
	var crds []string
	for _, file := range box.List() {
		if strings.Contains(file, crdYAMLpattern) {
			crds = append(crds, file)
		}
	}
	return crds
}
