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

package kafka

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	corev1 "k8s.io/api/core/v1"
)

const (
	enableEventsEnvKey = "ENABLE_EVENTS"

	// QuarkusBootstrapAppProp quarkus application property for setting kafka server
	QuarkusBootstrapAppProp = "kafka.bootstrap.servers"

	// SpringBootstrapAppProp spring boot application property for setting kafka server
	SpringBootstrapAppProp = "spring.kafka.bootstrap-servers"
)

// FetchInfraProperties provide application/env properties of infra that need to be ser in kogitoRuntime object
func (i *KafkaResource) FetchInfraProperties(instance *v1alpha1.KogitoInfra, runtimeType v1alpha1.RuntimeType) (appProps map[string]string, envProps []corev1.EnvVar) {
	appProps = map[string]string{}
	kafkaProperties := instance.Status.Kafka.KafkaProperties

	URI := kafkaProperties.ExternalURI
	if len(URI) > 0 {
		envProps = append(envProps, framework.CreateEnvVar(enableEventsEnvKey, "true"))
		if runtimeType == v1alpha1.SpringBootRuntimeType {
			appProps[SpringBootstrapAppProp] = URI
		} else {
			appProps[QuarkusBootstrapAppProp] = URI
		}
	} else {
		envProps = append(envProps, framework.CreateEnvVar(enableEventsEnvKey, "false"))
	}
	return
}
