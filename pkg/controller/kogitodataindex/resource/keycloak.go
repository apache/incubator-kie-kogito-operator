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

package resource

import (
	keycloakv1alpha1 "github.com/keycloak/keycloak-operator/pkg/apis/keycloak/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	userName = "kogito-data-index-viewer"
	password = "password"
	userRole = "confidential"

	keycloakClientName      = "kogito-data-index-keycloak-client"
	keycloakUserName        = "kogito-data-index-keycloak-user"
	clientID                = "kogito_data_index_service"
	secret                  = "secret"
	clientAuthenticatorType = "client-secret"
)

func newKeycloakUser(namespace string, realmLabels map[string]string) *keycloakv1alpha1.KeycloakUser {
	return &keycloakv1alpha1.KeycloakUser{
		ObjectMeta: v1.ObjectMeta{
			Name:      keycloakUserName,
			Namespace: namespace,
		},
		Spec: keycloakv1alpha1.KeycloakUserSpec{
			RealmSelector: &v1.LabelSelector{
				MatchLabels: realmLabels,
			},
			User: keycloakv1alpha1.KeycloakAPIUser{
				UserName:      userName,
				Enabled:       true,
				EmailVerified: false,
				RealmRoles: []string{
					userRole,
				},
				Credentials: []keycloakv1alpha1.KeycloakCredential{
					{
						Type:      "password",
						Value:     password,
						Temporary: false,
					},
				},
			},
		},
	}
}

func newKeycloakClient(namespace string, realmLabels map[string]string) *keycloakv1alpha1.KeycloakClient {
	return &keycloakv1alpha1.KeycloakClient{
		ObjectMeta: v1.ObjectMeta{
			Name:      keycloakClientName,
			Namespace: namespace,
		},
		Spec: keycloakv1alpha1.KeycloakClientSpec{
			RealmSelector: &v1.LabelSelector{
				MatchLabels: realmLabels,
			},
			Client: &keycloakv1alpha1.KeycloakAPIClient{
				ClientID:                  clientID,
				PublicClient:              false,
				BearerOnly:                false,
				ClientAuthenticatorType:   clientAuthenticatorType,
				Secret:                    secret,
				ServiceAccountsEnabled:    true,
				DirectAccessGrantsEnabled: true,
				RedirectUris:              []string{"*"},
				WebOrigins:                []string{"*"},
				Access: map[string]bool{
					"view":      true,
					"configure": true,
					"manage":    true,
				},
			},
		},
	}
}

// IsKeycloakReady checks if the Keycloak resources are provided and ready for Data Index
func IsKeycloakReady(instance *v1alpha1.KogitoDataIndex, client *client.Client) (bool, error) {
	if keycloak, err := infrastructure.GetKeycloakInstance(instance.Spec.KeycloakProperties.Keycloak, instance.Namespace, client); err != nil {
		return false, err
	} else if keycloak == nil || !keycloak.Status.Ready || keycloak.Status.Phase != keycloakv1alpha1.PhaseReconciling {
		return false, nil
	}

	if keycloakRealm, err := infrastructure.GetKeycloakRealmInstance(instance.Spec.KeycloakProperties.KeycloakRealm, instance.Namespace, client); err != nil {
		return false, err
	} else if keycloakRealm == nil || !keycloakRealm.Status.Ready || keycloakRealm.Status.Phase != keycloakv1alpha1.PhaseReconciling {
		return false, nil
	}

	if keycloakClient, err := infrastructure.GetKeycloakClientInstance(keycloakClientName, instance.Namespace, client); err != nil {
		return false, err
	} else if keycloakClient == nil || !keycloakClient.Status.Ready || keycloakClient.Status.Phase != keycloakv1alpha1.PhaseReconciling {
		return false, nil
	}

	if keycloakUser, err := infrastructure.GetKeycloakUserInstance(keycloakUserName, instance.Namespace, client); err != nil {
		return false, err
	} else if keycloakUser == nil || keycloakUser.Status.Phase != keycloakv1alpha1.PhaseReconciling {
		return false, nil
	}

	return true, nil
}
