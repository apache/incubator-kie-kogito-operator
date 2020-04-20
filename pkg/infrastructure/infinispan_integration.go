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
	"time"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// InfinispanSecretUsernameKey is the secret username key set in the linked secret
	InfinispanSecretUsernameKey = "username"
	// InfinispanSecretPasswordKey is the secret password key set in the linked secret
	InfinispanSecretPasswordKey = "password"

	infinispanEnvKeyUsername          string = "INFINISPAN_USERNAME"
	infinispanEnvKeyPassword          string = "INFINISPAN_PASSWORD"
	infinispanEnvKeyUseAuth           string = "INFINISPAN_USEAUTH"
	infinispanEnvKeyAuthRealm         string = "INFINISPAN_AUTHREALM"
	infinispanEnvKeySasl              string = "INFINISPAN_SASLMECHANISM"
	infinispanEnvKeyCredSecret        string = "INFINISPAN_CREDENTIAL_SECRET"
	infinispanEnvKeyServerList        string = "INFINISPAN_CLIENT_SERVER_LIST"
	infinispanEnvKeyQuarkusServerList string = "QUARKUS_INFINISPAN_CLIENT_SERVER_LIST"
)

// DeployInfinispanWithKogitoInfra deploys KogitoInfra with Infinispan enabled and sets the Infinispan credentials into the given InfinispanAware instance
// returns update = true if the instance needs to be updated, a duration for requeue and error != nil if something goes wrong
func DeployInfinispanWithKogitoInfra(instance v1alpha1.InfinispanAware, namespace string, cli *client.Client) (update bool, requeueAfter time.Duration, err error) {
	if instance == nil {
		return false, 0, nil
	}

	// Overrides any parameters not set
	if instance.GetInfinispanProperties().UseKogitoInfra {
		// ensure infra
		infra, ready, err := EnsureKogitoInfra(namespace, cli).WithInfinispan().Apply()
		if err != nil {
			return false, 0, err
		}

		uri, err := GetInfinispanServiceURI(cli, infra)
		if err != nil {
			return false, 0, err
		}
		if instance.GetInfinispanProperties().URI == uri &&
			instance.GetInfinispanProperties().Credentials.SecretName == infra.Status.Infinispan.CredentialSecret &&
			instance.GetInfinispanProperties().Credentials.PasswordKey == InfinispanSecretPasswordKey &&
			instance.GetInfinispanProperties().Credentials.UsernameKey == InfinispanSecretUsernameKey {
			return false, 0, nil
		}

		log.Debugf("Checking KogitoInfra status to make sure we are ready to use Infinispan. Status are: %s", infra.Status.Infinispan)
		if ready {
			log.Debug("Looks ok, we are ready to use Infinispan!")
			instance.SetInfinispanProperties(v1alpha1.InfinispanConnectionProperties{
				URI: uri,
				Credentials: v1alpha1.SecretCredentialsType{
					SecretName:  infra.Status.Infinispan.CredentialSecret,
					UsernameKey: InfinispanSecretUsernameKey,
					PasswordKey: InfinispanSecretPasswordKey,
				},
			})

			return true, 0, nil
		}
		log.Debug("KogitoInfra is not ready, requeue")
		// waiting for infinispan deployment
		return false, time.Second * 10, nil
	}

	// Ensure default values
	if instance.AreInfinispanPropertiesBlank() {
		instance.SetInfinispanProperties(v1alpha1.InfinispanConnectionProperties{
			UseKogitoInfra: true,
		})
		return true, 0, nil
	}

	return false, 0, nil
}

// FetchInfinispanCredentials fetches the secret defined in InfinispanAware instance
func FetchInfinispanCredentials(instance v1alpha1.InfinispanAware, namespace string, client *client.Client) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	if len(instance.GetInfinispanProperties().Credentials.SecretName) > 0 {
		secret.ObjectMeta = metav1.ObjectMeta{
			Name:      instance.GetInfinispanProperties().Credentials.SecretName,
			Namespace: namespace,
		}
		exist, err := kubernetes.ResourceC(client).Fetch(secret)
		if err != nil {
			return nil, err
		}
		if !exist {
			return nil, nil
		}
	}

	return secret, nil
}

// SetInfinispanVariables binds Infinispan properties in the container.
// If credentials are enabled, uses the given secret as a reference (use FetchInfinispanCredentials to get the secret)
func SetInfinispanVariables(infinispanProps v1alpha1.InfinispanConnectionProperties, secret *corev1.Secret, container *corev1.Container) {
	if &infinispanProps.Credentials != nil &&
		len(infinispanProps.Credentials.SecretName) > 0 &&
		secret != nil &&
		container != nil {

		usernameValue := InfinispanSecretUsernameKey
		if len(infinispanProps.Credentials.UsernameKey) > 0 {
			usernameValue = infinispanProps.Credentials.UsernameKey
		}
		framework.SetEnvVarFromSecret(infinispanEnvKeyUsername, usernameValue, secret, container)

		passwordValue := InfinispanSecretPasswordKey
		if len(infinispanProps.Credentials.PasswordKey) > 0 {
			passwordValue = infinispanProps.Credentials.PasswordKey
		}
		framework.SetEnvVarFromSecret(infinispanEnvKeyPassword, passwordValue, secret, container)

		framework.SetEnvVar(infinispanEnvKeyUseAuth, "true", container)
		framework.SetEnvVar(infinispanEnvKeyCredSecret, infinispanProps.Credentials.SecretName, container)
		if len(infinispanProps.SaslMechanism) == 0 {
			infinispanProps.SaslMechanism = v1alpha1.SASLPlain
		}
	} else {
		framework.SetEnvVar(infinispanEnvKeyUseAuth, "false", container)
	}
	if len(infinispanProps.AuthRealm) > 0 {
		framework.SetEnvVar(infinispanEnvKeyAuthRealm, infinispanProps.AuthRealm, container)
	}
	if len(infinispanProps.SaslMechanism) > 0 {
		framework.SetEnvVar(infinispanEnvKeySasl, string(infinispanProps.SaslMechanism), container)
	}
	if len(infinispanProps.URI) > 0 {
		framework.SetEnvVar(infinispanEnvKeyQuarkusServerList, infinispanProps.URI, container)
		framework.SetEnvVar(infinispanEnvKeyServerList, infinispanProps.URI, container)
	}
}
