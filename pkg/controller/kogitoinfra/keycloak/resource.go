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
	keycloakv1alpha1 "github.com/keycloak/keycloak-operator/pkg/apis/keycloak/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// keycloakMetricsExtension default extension enabled in Keycloak default installations
	keycloakMetricsExtension = "https://github.com/aerogear/keycloak-metrics-spi/releases/download/1.0.4/keycloak-metrics-spi-1.0.4.jar"
)

var log = logger.GetLogger("kogitokeycloak_resource")

func loadDeployedKeycloakInstance(cli *client.Client, name string, namespace string) (*keycloakv1alpha1.Keycloak, error) {
	log.Debug("fetching deployed kogito Keycloak instance")
	keycloakInstance := &keycloakv1alpha1.Keycloak{}
	if exits, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, keycloakInstance); err != nil {
		log.Error("Error occurs while fetching kogito Keycloak instance")
		return nil, err
	} else if !exits {
		log.Debug("Kogito Keycloak instance is not exists")
		return nil, nil
	} else {
		log.Debug("Kogito Keycloak instance found")
		return keycloakInstance, nil
	}
}

func createNewKeycloakInstance(cli *client.Client, name string, namespace string, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) (*keycloakv1alpha1.Keycloak, error) {
	log.Debug("Going to create kogito Keycloak instance")
	log.Debugf("Creating default resources for Keycloak installation for Kogito Infra on %s namespace", namespace)
	keycloakInstance := &keycloakv1alpha1.Keycloak{
		ObjectMeta: v1.ObjectMeta{Namespace: namespace, Name: name},
		Spec: keycloakv1alpha1.KeycloakSpec{
			Extensions:     []string{keycloakMetricsExtension},
			Instances:      1,
			ExternalAccess: keycloakv1alpha1.KeycloakExternalAccess{Enabled: true},
		},
	}
	if err := controllerutil.SetOwnerReference(instance, keycloakInstance, scheme); err != nil {
		return nil, err
	}
	if err := kubernetes.ResourceC(cli).Create(keycloakInstance); err != nil {
		log.Error("Error occurs while creating kogito Keycloak instance")
		return nil, err
	}
	log.Debug("Successfully created Kogito Keycloak instance")
	return keycloakInstance, nil
}
