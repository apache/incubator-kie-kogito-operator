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

package openshift

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	v1 "github.com/openshift/api/apps/v1"
)

// DeploymentConfigInterface common client for Deployment Config API interactions
type DeploymentConfigInterface interface {
	// RolloutLatest rolls out the latest deployment for the given DC
	RolloutLatest(dcName string, namespace string) (dc *v1.DeploymentConfig, err error)
}

func newDeploymentConfig(c *client.Client) DeploymentConfigInterface {
	client.MustEnsureClient(c)
	return &deploymentConfig{client: c}
}

type deploymentConfig struct {
	client *client.Client
}

func (d *deploymentConfig) RolloutLatest(dcName string, namespace string) (dc *v1.DeploymentConfig, err error) {
	request := &v1.DeploymentRequest{
		Name:   dcName,
		Latest: true,
		Force:  true,
	}
	if dc, err = d.client.DeploymentCli.DeploymentConfigs(namespace).Instantiate(dcName, request); err != nil {
		return
	}
	return
}
