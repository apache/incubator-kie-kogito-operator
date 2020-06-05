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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

type kafkaProperty int

const (
	appPropKafkaBootstrapServerList kafkaProperty = iota

	envVarKafkaBootstrapSuffix = "_BOOTSTRAP_SERVERS"
)

var (
	propertiesKafkaQuarkus = map[kafkaProperty]string{
		appPropKafkaBootstrapServerList: services.QuarkusBootstrapAppProp,
	}
	propertiesKafkaSpring = map[kafkaProperty]string{
		appPropKafkaBootstrapServerList: services.SpringBootstrapAppProp,
	}
)

// CreateKafkaProperties creates Kafka properties by reading information from the KogitoInfra
func CreateKafkaProperties(cli *client.Client, kogitoInfra *v1alpha1.KogitoInfra, kogitoApp *v1alpha1.KogitoApp) (envs []corev1.EnvVar, appProps map[string]string, err error) {
	if kogitoApp != nil &&
		(kogitoInfra != nil &&
			(len(kogitoInfra.Status.Kafka.Service) > 0 ||
				len(kogitoInfra.Status.Kafka.Name) > 0)) {
		uri, err := infrastructure.GetKafkaServiceURI(cli, kogitoInfra)
		if err != nil {
			return nil, nil, err
		}

		appProps = map[string]string{}

		vars := propertiesKafkaQuarkus
		if kogitoApp.Spec.Runtime == v1alpha1.SpringbootRuntimeType {
			vars = propertiesKafkaSpring
		}

		appProps[vars[appPropKafkaBootstrapServerList]] = uri
		// let's also add a secret feature that injects all _BOOTSTRAP_SERVERS env vars with the correct uri :p
		for _, env := range kogitoApp.Spec.KogitoServiceSpec.Envs {
			if strings.HasSuffix(env.Name, envVarKafkaBootstrapSuffix) {
				envs = append(envs, corev1.EnvVar{Name: env.Name, Value: uri})
			}
		}
	}
	return envs, appProps, nil
}
