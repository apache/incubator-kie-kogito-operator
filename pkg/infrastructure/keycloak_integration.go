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

package infrastructure

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"time"
)

const (
	profileEnv    = "QUARKUS_PROFILE"
	authServerEnv = "QUARKUS_OIDC_AUTH_SERVER_URL"
	clientEnv     = "QUARKUS_OIDC_CLIENT_ID"
	secretEnv     = "QUARKUS_OIDC_CREDENTIALS_SECRET"

	securityProfile = "keycloak"

	authPath = "/auth/realms/"
)

// DeployKeycloakWithKogitoInfra deploys KogitoInfra with Keycloak enabled
// returns update = true if the instance needs to be updated, a duration for requeue and error != nil if something goes wrong
func DeployKeycloakWithKogitoInfra(instance v1alpha1.KeycloakAware, namespace string, cli *client.Client) (update bool, requeueAfter time.Duration, err error) {
	if instance == nil {
		return false, 0, nil
	}

	// ensure infra
	infra, ready, err := EnsureKogitoInfra(namespace, cli).WithKeycloak().Apply()
	if err != nil {
		return false, 0, err
	}

	log.Debugf("Checking KogitoInfra status to make sure we are ready to use Keycloak. Status are: %s", infra.Status.Keycloak)
	if ready {
		keycloak, keycloakURL, err := GetKeycloakProperties(cli, infra)
		if err != nil {
			return false, 0, err
		}
		keycloakRealm, realmName, keycloakRealmLabels, err := GetKeycloakRealmProperties(cli, infra)
		if err != nil {
			return false, 0, err
		}
		if len(keycloakURL) > 0 && len(realmName) > 0 && len(keycloakRealmLabels) > 0 {
			if instance.GetKeycloakProperties().Keycloak == keycloak &&
				instance.GetKeycloakProperties().KeycloakRealm == keycloakRealm &&
				instance.GetKeycloakProperties().AuthServerURL == keycloakURL &&
				instance.GetKeycloakProperties().RealmName == realmName &&
				reflect.DeepEqual(instance.GetKeycloakProperties().Labels, keycloakRealmLabels) {
				return false, 0, nil
			}

			log.Debug("Looks ok, we are ready to use Keycloak!")
			instance.SetKeycloakProperties(v1alpha1.KeycloakConnectionProperties{
				Keycloak:      keycloak,
				KeycloakRealm: keycloakRealm,
				AuthServerURL: keycloakURL,
				RealmName:     realmName,
				Labels:        keycloakRealmLabels,
			})

			return true, 0, nil
		}
	}
	log.Debug("KogitoInfra is not ready, requeue")
	// waiting for infinispan deployment
	return false, time.Second * 10, nil
}

// SetKeycloakVariables binds Keycloak properties in the container.
func SetKeycloakVariables(keycloakProps v1alpha1.KeycloakConnectionProperties, clientID, secret string, container *corev1.Container) {
	if len(keycloakProps.AuthServerURL) > 0 && len(keycloakProps.RealmName) > 0 {
		profile := framework.GetEnvVarFromContainer(profileEnv, container)
		if len(profile) > 0 {
			profile = profile + " " + securityProfile
		} else {
			profile = securityProfile
		}
		framework.SetEnvVar(profileEnv, profile, container)
		framework.SetEnvVar(authServerEnv, keycloakProps.AuthServerURL+authPath+keycloakProps.RealmName, container)
		framework.SetEnvVar(clientEnv, clientID, container)
		framework.SetEnvVar(secretEnv, secret, container)
	}
}
