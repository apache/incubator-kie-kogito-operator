// Copyright 2020 Red Hat, Inc. and/or its affiliates
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

package keycloak

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/read"
	keycloakv1alpha1 "github.com/keycloak/keycloak-operator/pkg/apis/keycloak/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/api/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/prometheus/common/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

const (
	// InstanceName is the default instance name for Keycloak CR managed by Kogito
	InstanceName = "kogito-keycloak"
	// keycloakMetricsExtension default extension enabled in Keycloak default installations
	keycloakMetricsExtension = "https://github.com/aerogear/keycloak-metrics-spi/releases/download/1.0.4/keycloak-metrics-spi-1.0.4.jar"
)

// GetDeployedResources gets the resources deployed as is
func GetDeployedResources(kogitoInfra *v1alpha1.KogitoInfra, cli *client.Client) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	if infrastructure.IsKeycloakAvailable(cli) {
		reader := read.New(cli.ControlCli).WithNamespace(kogitoInfra.Namespace).WithOwnerObject(kogitoInfra)
		resources, err = reader.ListAll(&keycloakv1alpha1.KeycloakList{})
		if err != nil {
			log.Warn("Failed to list deployed objects. ", err)
			return nil, err
		}
	}

	return
}

// CreateRequiredResources creates the very basic resources to have Keycloak in the namespace
func CreateRequiredResources(kogitoInfra *v1alpha1.KogitoInfra) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	resources = make(map[reflect.Type][]resource.KubernetesResource, 1)
	if kogitoInfra.Spec.InstallKeycloak {
		log.Debugf("Creating default resources for Keycloak installation for Kogito Infra on %s namespace", kogitoInfra.Namespace)
		keycloak := &keycloakv1alpha1.Keycloak{
			ObjectMeta: v1.ObjectMeta{Namespace: kogitoInfra.Namespace, Name: InstanceName},
			Spec: keycloakv1alpha1.KeycloakSpec{
				Extensions:     []string{keycloakMetricsExtension},
				Instances:      1,
				ExternalAccess: keycloakv1alpha1.KeycloakExternalAccess{Enabled: true},
			},
		}
		resources[reflect.TypeOf(keycloakv1alpha1.Keycloak{})] = []resource.KubernetesResource{keycloak}
		log.Debugf("Requested objects created as %s", resources)
	}
	return
}
