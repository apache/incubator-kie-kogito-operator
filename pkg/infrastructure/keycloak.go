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
	keycloakv1alpha1 "github.com/keycloak/keycloak-operator/pkg/apis/keycloak/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"k8s.io/apimachinery/pkg/types"
)

const (
	keycloakServerGroup = "keycloak.org"

	// keycloakAppLabelKey label key to mark kogito app
	keycloakAppLabelKey = "app"
	// keycloakResourceLabelKey label key to mark security solution
	keycloakResourceLabelKey = "security"
)

// IsKeycloakAvailable checks if Strimzi CRD is available in the cluster
func IsKeycloakAvailable(client *client.Client) bool {
	return client.HasServerGroup(keycloakServerGroup)
}

// GetKeycloakProperties retrieves the properties of the Keycloak instance that is deployed by Kogito Infra
func GetKeycloakProperties(cli *client.Client, infra *v1alpha1.KogitoInfra) (name, URL string, err error) {
	if keycloak, err := GetKeycloakInstance(infra.Status.Keycloak.Name, infra.Namespace, cli); err != nil {
		return "", "", err
	} else if keycloak != nil &&
		keycloak.Status.Ready &&
		keycloak.Status.Phase == keycloakv1alpha1.PhaseReconciling &&
		len(keycloak.Status.InternalURL) > 0 {
		return keycloak.Name, keycloak.Status.InternalURL, nil
	}
	return "", "", nil
}

// GetKeycloakRealmProperties retrieves the properties of the Keycloak Realm instance that is deployed by Kogito Infra
func GetKeycloakRealmProperties(cli *client.Client, infra *v1alpha1.KogitoInfra) (name, realmName string, labels map[string]string, err error) {
	if keycloakRealm, err := GetKeycloakRealmInstance(infra.Status.Keycloak.RealmStatus.Name, infra.Namespace, cli); err != nil {
		return "", "", nil, err
	} else if keycloakRealm != nil &&
		keycloakRealm.Status.Ready &&
		keycloakRealm.Status.Phase == keycloakv1alpha1.PhaseReconciling {
		return keycloakRealm.Name,
			keycloakRealm.Spec.Realm.Realm,
			map[string]string{
				keycloakAppLabelKey:      keycloakRealm.Labels[keycloakAppLabelKey],
				keycloakResourceLabelKey: keycloakRealm.Labels[keycloakResourceLabelKey],
			},
			nil
	}
	return "", "", nil, nil
}

// GetKeycloakInstance fetches the Keycloak instance of the given name
func GetKeycloakInstance(name string, namespace string, client *client.Client) (*keycloakv1alpha1.Keycloak, error) {
	keycloak := &keycloakv1alpha1.Keycloak{}
	if exists, err := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, keycloak); err != nil {
		return nil, err
	} else if exists {
		return keycloak, nil
	}
	return nil, nil
}

// GetKeycloakRealmInstance fetches the Keycloak Realm instance of the given name
func GetKeycloakRealmInstance(name string, namespace string, client *client.Client) (*keycloakv1alpha1.KeycloakRealm, error) {
	keycloakRealm := &keycloakv1alpha1.KeycloakRealm{}
	if exists, err := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, keycloakRealm); err != nil {
		return nil, err
	} else if exists {
		return keycloakRealm, nil
	}
	return nil, nil
}

// GetKeycloakClientInstance fetches the Keycloak Client instance of the given name
func GetKeycloakClientInstance(name string, namespace string, client *client.Client) (*keycloakv1alpha1.KeycloakClient, error) {
	keycloakClient := &keycloakv1alpha1.KeycloakClient{}
	if exists, err := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, keycloakClient); err != nil {
		return nil, err
	} else if exists {
		return keycloakClient, nil
	}
	return nil, nil
}

// GetKeycloakUserInstance fetches the Keycloak User instance of the given name
func GetKeycloakUserInstance(name string, namespace string, client *client.Client) (*keycloakv1alpha1.KeycloakUser, error) {
	keycloakUser := &keycloakv1alpha1.KeycloakUser{}
	if exists, err := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, keycloakUser); err != nil {
		return nil, err
	} else if exists {
		return keycloakUser, nil
	}
	return nil, nil
}
