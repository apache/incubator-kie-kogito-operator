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
	infinispan "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// keys for infinispan vars definition

	// appPropInfinispanServerList application property for setting infinispan server
	appPropInfinispanServerList int = iota
	// appPropInfinispanUseAuth application property for enabling infinispan authentication
	appPropInfinispanUseAuth
	// appPropInfinispanSaslMechanism application property for setting infinispan SASL mechanism
	appPropInfinispanSaslMechanism
	// appPropInfinispanAuthRealm application property for setting infinispan auth realm
	appPropInfinispanAuthRealm
	// envVarInfinispanUser environment variable for setting infinispan username
	envVarInfinispanUser
	// envVarInfinispanPassword environment variable for setting infinispan password
	envVarInfinispanPassword
	infinispanEnvKeyCredSecret = "INFINISPAN_CREDENTIAL_SECRET"
	enablePersistenceEnvKey    = "ENABLE_PERSISTENCE"
)

var (
	//Infinispan variables for the KogitoInfra deployed infrastructure.
	//For Quarkus: https://quarkus.io/guides/infinispan-client#quarkus-infinispan-client_configuration
	//For Spring: https://github.com/infinispan/infinispan-spring-boot/blob/master/infinispan-spring-boot-starter-remote/src/test/resources/test-application.properties

	// propertiesInfinispanQuarkus infinispan properties for quarkus runtime
	propertiesInfinispanQuarkus = map[int]string{
		appPropInfinispanServerList:    "quarkus.infinispan-client.server-list",
		appPropInfinispanUseAuth:       "quarkus.infinispan-client.use-auth",
		appPropInfinispanSaslMechanism: "quarkus.infinispan-client.sasl-mechanism",
		appPropInfinispanAuthRealm:     "quarkus.infinispan-client.auth-realm",

		envVarInfinispanUser:     "QUARKUS_INFINISPAN_CLIENT_AUTH_USERNAME",
		envVarInfinispanPassword: "QUARKUS_INFINISPAN_CLIENT_AUTH_PASSWORD",
	}
	// propertiesInfinispanSpring infinispan properties for spring boot runtime
	propertiesInfinispanSpring = map[int]string{
		appPropInfinispanServerList:    "infinispan.remote.server-list",
		appPropInfinispanUseAuth:       "infinispan.remote.use-auth",
		appPropInfinispanSaslMechanism: "infinispan.remote.sasl-mechanism",
		appPropInfinispanAuthRealm:     "infinispan.remote.auth-realm",

		envVarInfinispanUser:     "INFINISPAN_REMOTE_AUTH_USERNAME",
		envVarInfinispanPassword: "INFINISPAN_REMOTE_AUTH_PASSWORD",
	}
)

func getInfinispanSecretEnvVars(cli *client.Client, infinispanInstance *infinispan.Infinispan, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) ([]corev1.EnvVar, error) {
	var envProps []corev1.EnvVar

	customInfinispanSecret, resultErr := loadCustomKogitoInfinispanSecret(cli, instance.Namespace)
	if resultErr != nil {
		return nil, resultErr
	}

	if customInfinispanSecret == nil {
		customInfinispanSecret, resultErr = createCustomKogitoInfinispanSecret(cli, instance.Namespace, infinispanInstance, instance, scheme)
		if resultErr != nil {
			return nil, resultErr
		}
	}

	envProps = append(envProps, framework.CreateEnvVar(enablePersistenceEnvKey, "true"))
	secretName := customInfinispanSecret.Name
	envProps = append(envProps, framework.CreateEnvVar(infinispanEnvKeyCredSecret, secretName))
	envProps = append(envProps, framework.CreateSecretEnvVar(propertiesInfinispanSpring[envVarInfinispanUser], secretName, infrastructure.InfinispanSecretUsernameKey))
	envProps = append(envProps, framework.CreateSecretEnvVar(propertiesInfinispanQuarkus[envVarInfinispanUser], secretName, infrastructure.InfinispanSecretUsernameKey))
	envProps = append(envProps, framework.CreateSecretEnvVar(propertiesInfinispanSpring[envVarInfinispanPassword], secretName, infrastructure.InfinispanSecretPasswordKey))
	envProps = append(envProps, framework.CreateSecretEnvVar(propertiesInfinispanQuarkus[envVarInfinispanPassword], secretName, infrastructure.InfinispanSecretPasswordKey))
	return envProps, nil
}

func getInfinispanAppProps(cli *client.Client, name string, namespace string) (map[string]string, error) {
	appProps := map[string]string{}

	infinispanURI, resultErr := infrastructure.FetchKogitoInfinispanInstanceURI(cli, name, namespace)
	if resultErr != nil {
		return nil, resultErr
	}

	appProps[propertiesInfinispanSpring[appPropInfinispanUseAuth]] = "true"
	appProps[propertiesInfinispanQuarkus[appPropInfinispanUseAuth]] = "true"
	if len(infinispanURI) > 0 {
		appProps[propertiesInfinispanSpring[appPropInfinispanServerList]] = infinispanURI
		appProps[propertiesInfinispanQuarkus[appPropInfinispanServerList]] = infinispanURI
	}
	return appProps, nil
}
