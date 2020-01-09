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
	oappsv1 "github.com/openshift/api/apps/v1"
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

// getKogitoAppsDCs gets all dcs owned by KogitoApps within the given namespace
func getKogitoAppsDCs(namespace string, cli *client.Client) ([]oappsv1.DeploymentConfig, error) {
	var kdcs []oappsv1.DeploymentConfig
	kogitoApps := &v1alpha1.KogitoAppList{}
	if err := kubernetes.ResourceC(cli).ListWithNamespace(namespace, kogitoApps); err != nil {
		return nil, err
	}
	log.Debugf("Found %d KogitoApps in the namespace '%s' ", len(kogitoApps.Items), namespace)
	if len(kogitoApps.Items) == 0 {
		return kdcs, nil
	}
	dcs := &oappsv1.DeploymentConfigList{}
	if err := kubernetes.ResourceC(cli).ListWithNamespace(namespace, dcs); err != nil {
		return nil, err
	}
	log.Debug("Looking for DeploymentConfigs owned by KogitoApps")
	for _, dc := range dcs.Items {
		for _, owner := range dc.OwnerReferences {
			for _, app := range kogitoApps.Items {
				if owner.UID == app.UID {
					kdcs = append(kdcs, dc)
					break
				}
			}
		}
	}
	return kdcs, nil
}
