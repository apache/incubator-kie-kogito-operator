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

package infrastructure

import (
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/operator"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckKogitoOperatorExists checks whether Kogito operator is existing and running.
// If it is existing but not running, it returns true and an error
func CheckKogitoOperatorExists(kubeCli *client.Client, namespace string) (bool, error) {
	return CheckOperatorExists(kubeCli, namespace, operator.Name)
}

// CheckOperatorExists checks whether operator is existing and running.
// If it is existing but not running, it returns true and an error
func CheckOperatorExists(kubeCli *client.Client, namespace, operatorName string) (bool, error) {
	log.Debugf("Checking if %s Operator is deployed in namespace %s", operatorName, namespace)
	operatorDeployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operatorName,
			Namespace: namespace,
		},
	}

	if exists, err := kubernetes.ResourceC(kubeCli).Fetch(operatorDeployment); err != nil {
		return false, fmt.Errorf("Error while trying to look for %s Operator installation: %s ", operatorName, err)
	} else if !exists {
		return false, nil
	}

	if operatorDeployment.Status.AvailableReplicas == 0 {
		return true, fmt.Errorf("%s Operator seems to be created in the namespace '%s', but there's no available pods replicas deployed ", operatorName, namespace)
	}

	return true, nil
}
