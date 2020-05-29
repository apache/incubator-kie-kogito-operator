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

package services

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const (
	// vars for services that relies on infinispan image module, keeping for backward compatibility, should be removed:
	// https://issues.redhat.com/browse/KOGITO-2046
	// Deprecated
	infinispanEnvKeyUsername string = "INFINISPAN_USERNAME"
	// Deprecated
	infinispanEnvKeyPassword string = "INFINISPAN_PASSWORD"
	// Deprecated
	infinispanEnvKeyUseAuth string = "INFINISPAN_USEAUTH"
	// Deprecated
	infinispanEnvKeyAuthRealm string = "INFINISPAN_AUTHREALM"
	// Deprecated
	infinispanEnvKeySasl string = "INFINISPAN_SASLMECHANISM"
	// Deprecated
	infinispanEnvKeyServerList string = "INFINISPAN_CLIENT_SERVER_LIST"

	infinispanEnvKeyCredSecret string = "INFINISPAN_CREDENTIAL_SECRET"

	// keys for infinispan vars definition

	envVarInfinispanServerList    = "SERVER_LIST"
	envVarInfinispanUseAuth       = "USE_AUTH"
	envVarInfinispanUser          = "USERNAME"
	envVarInfinispanPassword      = "PASSWORD"
	envVarInfinispanSaslMechanism = "SASL_MECHANISM"
	envVarInfinispanAuthRealm     = "AUTH_REALM"
)

var (
	/*
		Infinispan variables for the KogitoInfra deployed infrastructure.
		For Quarkus: https://quarkus.io/guides/infinispan-client#quarkus-infinispan-client_configuration
		For Spring: https://github.com/infinispan/infinispan-spring-boot/blob/master/infinispan-spring-boot-starter-remote/src/test/resources/test-application.properties
	*/

	envVarInfinispanQuarkus = map[string]string{
		envVarInfinispanServerList:    "QUARKUS_INFINISPAN_CLIENT_SERVER_LIST",
		envVarInfinispanUseAuth:       "QUARKUS_INFINISPAN_CLIENT_USE_AUTH",
		envVarInfinispanUser:          "QUARKUS_INFINISPAN_CLIENT_AUTH_USERNAME",
		envVarInfinispanPassword:      "QUARKUS_INFINISPAN_CLIENT_AUTH_PASSWORD",
		envVarInfinispanSaslMechanism: "QUARKUS_INFINISPAN_CLIENT_SASL_MECHANISM",
		envVarInfinispanAuthRealm:     "QUARKUS_INFINISPAN_CLIENT_AUTH_REALM",
	}
	envVarInfinispanSpring = map[string]string{
		envVarInfinispanServerList:    "INFINISPAN_REMOTE_SERVER_LIST",
		envVarInfinispanUseAuth:       "INFINISPAN_REMOTE_USE_AUTH",
		envVarInfinispanUser:          "INFINISPAN_REMOTE_AUTH_USERNAME",
		envVarInfinispanPassword:      "INFINISPAN_REMOTE_AUTH_PASSWORD",
		envVarInfinispanSaslMechanism: "INFINISPAN_REMOTE_SASL_MECHANISM",
		envVarInfinispanAuthRealm:     "INFINISPAN_REMOTE_AUTH_REALM",
	}
)

// setInfinispanVariables binds Infinispan properties in the container.
// If credentials are enabled, uses the given secret as a reference (use FetchInfinispanCredentials to get the secret)
func setInfinispanVariables(runtime v1alpha1.RuntimeType, infinispanProps v1alpha1.InfinispanConnectionProperties, secret *corev1.Secret, container *corev1.Container) {
	vars := envVarInfinispanQuarkus
	if runtime == v1alpha1.SpringbootRuntimeType {
		vars = envVarInfinispanSpring
	}

	if &infinispanProps.Credentials != nil &&
		len(infinispanProps.Credentials.SecretName) > 0 &&
		secret != nil &&
		container != nil {

		usernameValue := infrastructure.InfinispanSecretUsernameKey
		if len(infinispanProps.Credentials.UsernameKey) > 0 {
			usernameValue = infinispanProps.Credentials.UsernameKey
		}
		framework.SetEnvVarFromSecret(infinispanEnvKeyUsername, usernameValue, secret, container)
		framework.SetEnvVarFromSecret(vars[envVarInfinispanUser], usernameValue, secret, container)

		passwordValue := infrastructure.InfinispanSecretPasswordKey
		if len(infinispanProps.Credentials.PasswordKey) > 0 {
			passwordValue = infinispanProps.Credentials.PasswordKey
		}
		framework.SetEnvVarFromSecret(infinispanEnvKeyPassword, passwordValue, secret, container)
		framework.SetEnvVarFromSecret(vars[envVarInfinispanPassword], passwordValue, secret, container)

		framework.SetEnvVar(infinispanEnvKeyUseAuth, "true", container)
		framework.SetEnvVar(vars[envVarInfinispanUseAuth], "true", container)

		framework.SetEnvVar(infinispanEnvKeyCredSecret, infinispanProps.Credentials.SecretName, container)
		if len(infinispanProps.SaslMechanism) == 0 {
			infinispanProps.SaslMechanism = v1alpha1.SASLPlain
		}
	} else {
		framework.SetEnvVar(infinispanEnvKeyUseAuth, "false", container)
		framework.SetEnvVar(vars[envVarInfinispanUseAuth], "false", container)
	}
	if len(infinispanProps.AuthRealm) > 0 {
		framework.SetEnvVar(infinispanEnvKeyAuthRealm, infinispanProps.AuthRealm, container)
		framework.SetEnvVar(vars[envVarInfinispanAuthRealm], infinispanProps.AuthRealm, container)
	}
	if len(infinispanProps.SaslMechanism) > 0 {
		framework.SetEnvVar(infinispanEnvKeySasl, string(infinispanProps.SaslMechanism), container)
		framework.SetEnvVar(vars[envVarInfinispanSaslMechanism], string(infinispanProps.SaslMechanism), container)
	}
	if len(infinispanProps.URI) > 0 {
		framework.SetEnvVar(infinispanEnvKeyServerList, infinispanProps.URI, container)
		framework.SetEnvVar(vars[envVarInfinispanServerList], infinispanProps.URI, container)
	}
}

// fetchInfinispanCredentials fetches the secret defined in InfinispanAware instance
func fetchInfinispanCredentials(instance v1alpha1.InfinispanAware, namespace string, client *client.Client) (*corev1.Secret, error) {
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

// deployInfinispanWithKogitoInfra deploys KogitoInfra with Infinispan enabled and sets the Infinispan credentials into the given InfinispanAware instance
// returns update = true if the instance needs to be updated, a duration for requeue and error != nil if something goes wrong
func deployInfinispanWithKogitoInfra(instance v1alpha1.InfinispanAware, namespace string, cli *client.Client) (update bool, requeueAfter time.Duration, err error) {
	if instance == nil {
		return false, 0, nil
	}

	// Overrides any parameters not set
	if instance.GetInfinispanProperties().UseKogitoInfra {
		// ensure infra
		infra, ready, err := infrastructure.EnsureKogitoInfra(namespace, cli).WithInfinispan().Apply()
		if err != nil {
			return false, 0, err
		}

		uri, err := infrastructure.GetInfinispanServiceURI(cli, infra)
		if err != nil {
			return false, 0, err
		}
		if instance.GetInfinispanProperties().URI == uri &&
			instance.GetInfinispanProperties().Credentials.SecretName == infra.Status.Infinispan.CredentialSecret &&
			instance.GetInfinispanProperties().Credentials.PasswordKey == infrastructure.InfinispanSecretPasswordKey &&
			instance.GetInfinispanProperties().Credentials.UsernameKey == infrastructure.InfinispanSecretUsernameKey {
			return false, 0, nil
		}

		log.Debugf("Checking KogitoInfra status to make sure we are ready to use Infinispan. Status are: %s", infra.Status.Infinispan)
		if ready {
			log.Debug("Looks ok, we are ready to use Infinispan!")
			instance.SetInfinispanProperties(v1alpha1.InfinispanConnectionProperties{
				URI: uri,
				Credentials: v1alpha1.SecretCredentialsType{
					SecretName:  infra.Status.Infinispan.CredentialSecret,
					UsernameKey: infrastructure.InfinispanSecretUsernameKey,
					PasswordKey: infrastructure.InfinispanSecretPasswordKey,
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
