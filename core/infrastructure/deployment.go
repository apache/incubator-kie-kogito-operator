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
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

// DeploymentHandler ...
type DeploymentHandler interface {
	FetchDeployment(key types.NamespacedName) (*appsv1.Deployment, error)
	FetchDeploymentList(namespace string) (*appsv1.DeploymentList, error)
	MustFetchDeployment(key types.NamespacedName) (*appsv1.Deployment, error)
	IsDeploymentAvailable(key types.NamespacedName) (bool, error)
}

type deploymentHandler struct {
	client *client.Client
	log    logger.Logger
}

// NewDeploymentHandler ...
func NewDeploymentHandler(client *client.Client, log logger.Logger) DeploymentHandler {
	return &deploymentHandler{
		client: client,
		log:    log,
	}
}

func (d *deploymentHandler) FetchDeployment(key types.NamespacedName) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	_, err := kubernetes.ResourceC(d.client).FetchWithKey(key, deployment)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func (d *deploymentHandler) FetchDeploymentList(namespace string) (*appsv1.DeploymentList, error) {
	dcs := &appsv1.DeploymentList{}
	if err := kubernetes.ResourceC(d.client).ListWithNamespace(namespace, dcs); err != nil {
		return nil, err
	}
	return dcs, nil
}

func (d *deploymentHandler) MustFetchDeployment(key types.NamespacedName) (*appsv1.Deployment, error) {
	deployment, err := d.FetchDeployment(key)
	if err != nil {
		return nil, err
	} else if deployment == nil {
		return nil, fmt.Errorf("deployment not found with name %s in namespace %s", key.Name, key.Namespace)
	}
	return deployment, nil
}

// IsDeploymentAvailable verifies if the Deployment resource from the given KogitoService has replicas available
func (d *deploymentHandler) IsDeploymentAvailable(key types.NamespacedName) (bool, error) {
	deployment, err := d.FetchDeployment(key)
	if err != nil {
		return false, err
	} else if deployment == nil {
		return false, nil
	}
	return true, nil
}
