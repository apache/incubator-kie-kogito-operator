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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	// ConfigMapProtoBufEnabledLabelKey label key used by configMaps that are meant to hold protobuf files
	ConfigMapProtoBufEnabledLabelKey = "kogito-protobuf"
)

// GetProtoBufConfigMaps will return every configMap labeled as "protobuf=true" in the given namespace
func GetProtoBufConfigMaps(namespace string, cli *client.Client) (*v1.ConfigMapList, error) {
	cms := &v1.ConfigMapList{}
	if err := kubernetes.ResourceC(cli).ListWithNamespaceAndLabel(namespace, cms, map[string]string{ConfigMapProtoBufEnabledLabelKey: "true"}); err != nil {
		return nil, err
	}
	return cms, nil
}

// getKogitoRuntimeDeployments gets all dcs owned by KogitoRuntime services within the given namespace
func getKogitoRuntimeDeployments(namespace string, cli *client.Client) ([]appsv1.Deployment, error) {
	var kdcs []appsv1.Deployment
	kogitoRuntimeServices := &v1alpha1.KogitoRuntimeList{}
	if err := kubernetes.ResourceC(cli).ListWithNamespace(namespace, kogitoRuntimeServices); err != nil {
		return nil, err
	}
	log.Debugf("Found %d KogitoRuntime services in the namespace '%s' ", len(kogitoRuntimeServices.Items), namespace)
	if len(kogitoRuntimeServices.Items) == 0 {
		return kdcs, nil
	}
	dcs := &appsv1.DeploymentList{}
	if err := kubernetes.ResourceC(cli).ListWithNamespace(namespace, dcs); err != nil {
		return nil, err
	}
	log.Debug("Looking for DeploymentConfigs owned by KogitoRuntime")
	for _, dc := range dcs.Items {
		for _, owner := range dc.OwnerReferences {
			for _, app := range kogitoRuntimeServices.Items {
				if owner.UID == app.UID {
					kdcs = append(kdcs, dc)
					break
				}
			}
		}
	}
	return kdcs, nil
}
