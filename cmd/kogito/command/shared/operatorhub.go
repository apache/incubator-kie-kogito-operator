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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"strings"
)

const (
	defaultOperatorPackageName   = "kogito-operator"
	communityOperatorSource      = "community-operators"
	operatorMarketplaceNamespace = "openshift-marketplace"
)

// isOperatorAvailableInOperatorHub will check if the Kogito Operator is available in OperatorHub (on OpenShift)
func isOperatorAvailableInOperatorHub(kubeCli *client.Client) (bool, error) {
	log := context.GetDefaultLogger()
	log.Debug("Trying to find if Kogito Operator is available in the OperatorHub")
	operatorSource := &v1.OperatorSource{
		ObjectMeta: v12.ObjectMeta{
			Name:      communityOperatorSource,
			Namespace: operatorMarketplaceNamespace,
		},
	}
	exists, err := kubernetes.ResourceC(kubeCli).Fetch(operatorSource)
	if err != nil {
		return false, err
	}

	log.Debugf("Finishing fetch the OperatorHub for Kogito Operator in namespace %s", operatorMarketplaceNamespace)
	log.Debugf("OperatorSource named as %s created at %s", operatorSource.Name, operatorSource.CreationTimestamp)
	log.Debugf("OperatorSource %s has the following packages: %s", operatorSource.Name, operatorSource.Status.Packages)
	if !exists {
		return false, nil
	}

	return strings.Contains(operatorSource.Status.Packages, defaultOperatorPackageName), nil
}
