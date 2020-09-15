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

package infinispan

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	corev1 "k8s.io/api/core/v1"
)

const (
	// keys for infinispan vars definition

	// AppPropInfinispanServerList application property for setting infinispan server
	AppPropInfinispanServerList int = iota
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

	infinispanEnvKeyCredSecret = "INFINISPAN_CREDENTIAL_SECRET"
	enablePersistenceEnvKey    = "ENABLE_PERSISTENCE"
)

var (
	//Infinispan variables for the KogitoInfra deployed infrastructure.
	//For Quarkus: https://quarkus.io/guides/infinispan-client#quarkus-infinispan-client_configuration
	//For Spring: https://github.com/infinispan/infinispan-spring-boot/blob/master/infinispan-spring-boot-starter-remote/src/test/resources/test-application.properties

	// PropertiesInfinispanQuarkus infinispan properties for quarkus runtime
	PropertiesInfinispanQuarkus = map[int]string{
		AppPropInfinispanServerList:    "quarkus.infinispan-client.server-list",
		AppPropInfinispanUseAuth:       "quarkus.infinispan-client.use-auth",
		AppPropInfinispanSaslMechanism: "quarkus.infinispan-client.sasl-mechanism",
		AppPropInfinispanAuthRealm:     "quarkus.infinispan-client.auth-realm",

		EnvVarInfinispanUser:     "QUARKUS_INFINISPAN_CLIENT_AUTH_USERNAME",
		EnvVarInfinispanPassword: "QUARKUS_INFINISPAN_CLIENT_AUTH_PASSWORD",
	}
	// PropertiesInfinispanSpring infinispan properties for spring boot runtime
	PropertiesInfinispanSpring = map[int]string{
		AppPropInfinispanServerList:    "infinispan.remote.server-list",
		AppPropInfinispanUseAuth:       "infinispan.remote.use-auth",
		AppPropInfinispanSaslMechanism: "infinispan.remote.sasl-mechanism",
		AppPropInfinispanAuthRealm:     "infinispan.remote.auth-realm",

		EnvVarInfinispanUser:     "INFINISPAN_REMOTE_AUTH_USERNAME",
		EnvVarInfinispanPassword: "INFINISPAN_REMOTE_AUTH_PASSWORD",
	}
)

// FetchInfraProperties provide application/env properties of infra that need to be set in the KogitoRuntime object
func (i *InfraResource) FetchInfraProperties(instance *v1alpha1.KogitoInfra, runtimeType v1alpha1.RuntimeType) (appProps map[string]string, envProps []corev1.EnvVar) {
	log.Debugf("going to fetch infinispan infra properties for given kogito infra instance : %s", instance.Name)

	appProps = map[string]string{}
	vars := PropertiesInfinispanQuarkus
	if runtimeType == v1alpha1.SpringBootRuntimeType {
		vars = PropertiesInfinispanSpring
	}

	envProps = append(envProps, framework.CreateEnvVar(enablePersistenceEnvKey, "true"))
	infinispanProps := instance.Status.InfinispanProperties
	secretName := infinispanProps.Credentials.SecretName
	if len(secretName) > 0 {
		envProps = append(envProps, framework.CreateEnvVar(infinispanEnvKeyCredSecret, secretName))
		envProps = append(envProps, framework.CreateSecretEnvVar(vars[EnvVarInfinispanUser], secretName, infinispanProps.Credentials.UsernameKey))
		envProps = append(envProps, framework.CreateSecretEnvVar(vars[EnvVarInfinispanPassword], secretName, infinispanProps.Credentials.PasswordKey))

		appProps[vars[AppPropInfinispanUseAuth]] = "true"
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
		appProps[vars[AppPropInfinispanServerList]] = infinispanProps.URI
	}
	return
}
