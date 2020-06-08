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

// InfinispanProperty type for infinispan properties
type InfinispanProperty int

const (
	// keys for infinispan vars definition

	// AppPropInfinispanServerList application property for setting infinispan server
	AppPropInfinispanServerList InfinispanProperty = iota
	// AppPropInfinispanUseAuth application property for enabling infinispan authentication
	AppPropInfinispanUseAuth
	// AppPropInfinispanSaslMechanism application property for setting infinispan SASL mechanism
	AppPropInfinispanSaslMechanism
	// AppPropInfinispanAuthRealm application property for setting infinispan auth realm
	AppPropInfinispanAuthRealm
	// EnvVarInfinispanUser environment variable for setting infinispan username
	EnvVarInfinispanUser
	// EnvVarInfinispanPassword environment variable for setting infinispan password
	EnvVarInfinispanPassword

	infinispanEnvKeyCredSecret string = "INFINISPAN_CREDENTIAL_SECRET"

	// Deprecated, keep it for Job Service scripts
	infinispanEnvKeyServerList string = "INFINISPAN_CLIENT_SERVER_LIST"
	// Deprecated, keep it for Data Index and Job Service scripts
	infinispanEnvKeyUsername string = "INFINISPAN_USERNAME"
	// Deprecated, keep it for Data Index and Job Service scripts
	infinispanEnvKeyPassword string = "INFINISPAN_PASSWORD"
)

var (
	/*
		Infinispan variables for the KogitoInfra deployed infrastructure.
		For Quarkus: https://quarkus.io/guides/infinispan-client#quarkus-infinispan-client_configuration
		For Spring: https://github.com/infinispan/infinispan-spring-boot/blob/master/infinispan-spring-boot-starter-remote/src/test/resources/test-application.properties
	*/

	// PropertiesInfinispanQuarkus infinispan properties for quarkus runtime
	PropertiesInfinispanQuarkus = map[InfinispanProperty]string{
		AppPropInfinispanServerList:    "quarkus.infinispan-client.server-list",
		AppPropInfinispanUseAuth:       "quarkus.infinispan-client.use-auth",
		AppPropInfinispanSaslMechanism: "quarkus.infinispan-client.sasl-mechanism",
		AppPropInfinispanAuthRealm:     "quarkus.infinispan-client.auth-realm",

		EnvVarInfinispanUser:     "QUARKUS_INFINISPAN_CLIENT_AUTH_USERNAME",
		EnvVarInfinispanPassword: "QUARKUS_INFINISPAN_CLIENT_AUTH_PASSWORD",
	}
	// PropertiesInfinispanSpring infinispan properties for spring boot runtime
	PropertiesInfinispanSpring = map[InfinispanProperty]string{
		AppPropInfinispanServerList:    "infinispan.remote.server-list",
		AppPropInfinispanUseAuth:       "infinispan.remote.use-auth",
		AppPropInfinispanSaslMechanism: "infinispan.remote.sasl-mechanism",
		AppPropInfinispanAuthRealm:     "infinispan.remote.auth-realm",

		EnvVarInfinispanUser:     "INFINISPAN_REMOTE_AUTH_USERNAME",
		EnvVarInfinispanPassword: "INFINISPAN_REMOTE_AUTH_PASSWORD",
	}
)

// setInfinispanVariables binds Infinispan properties in the container.
// If credentials are enabled, uses the given secret as a reference (use FetchInfinispanCredentials to get the secret)
func setInfinispanVariables(runtime v1alpha1.RuntimeType, infinispanProps *v1alpha1.InfinispanConnectionProperties, secret *corev1.Secret, container *corev1.Container, appProps map[string]string) {
	vars := PropertiesInfinispanQuarkus
	if runtime == v1alpha1.SpringbootRuntimeType {
		vars = PropertiesInfinispanSpring
	}

	if infinispanProps != nil &&
		len(infinispanProps.Credentials.SecretName) > 0 &&
		secret != nil &&
		container != nil {

		usernameValue := infrastructure.InfinispanSecretUsernameKey
		if len(infinispanProps.Credentials.UsernameKey) > 0 {
			usernameValue = infinispanProps.Credentials.UsernameKey
		}
		framework.SetEnvVarFromSecret(vars[EnvVarInfinispanUser], usernameValue, secret, container)
		framework.SetEnvVarFromSecret(infinispanEnvKeyUsername, usernameValue, secret, container)

		passwordValue := infrastructure.InfinispanSecretPasswordKey
		if len(infinispanProps.Credentials.PasswordKey) > 0 {
			passwordValue = infinispanProps.Credentials.PasswordKey
		}
		framework.SetEnvVarFromSecret(vars[EnvVarInfinispanPassword], passwordValue, secret, container)
		framework.SetEnvVarFromSecret(infinispanEnvKeyPassword, passwordValue, secret, container)

		appProps[vars[AppPropInfinispanUseAuth]] = "true"

		framework.SetEnvVar(infinispanEnvKeyCredSecret, infinispanProps.Credentials.SecretName, container)
		if len(infinispanProps.SaslMechanism) == 0 {
			infinispanProps.SaslMechanism = v1alpha1.SASLPlain
		}
	} else {
		appProps[vars[AppPropInfinispanUseAuth]] = "false"
	}
	if len(infinispanProps.AuthRealm) > 0 {
		appProps[vars[AppPropInfinispanAuthRealm]] = infinispanProps.AuthRealm
	}
	if len(infinispanProps.SaslMechanism) > 0 {
		appProps[vars[AppPropInfinispanSaslMechanism]] = string(infinispanProps.SaslMechanism)
	}
	if len(infinispanProps.URI) > 0 {
		framework.SetEnvVar(infinispanEnvKeyServerList, infinispanProps.URI, container)
		appProps[vars[AppPropInfinispanServerList]] = infinispanProps.URI
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
