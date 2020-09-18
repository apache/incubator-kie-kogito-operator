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
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	corev1 "k8s.io/api/core/v1"
)

const (
	enableEventsEnvKey = "ENABLE_EVENTS"

	// quarkusBootstrapAppProp quarkus application property for setting kafka server
	quarkusBootstrapAppProp = "kafka.bootstrap.servers"

	// springBootstrapAppProp spring boot application property for setting kafka server
	springBootstrapAppProp = "spring.kafka.bootstrap-servers"
)

func getKafkaEnvVars(kafkaInstance *kafkabetav1.Kafka) []corev1.EnvVar {
	kafkaURI := infrastructure.ResolveKafkaServerURI(kafkaInstance)
	var envProps []corev1.EnvVar
	if len(kafkaURI) > 0 {
		envProps = append(envProps, framework.CreateEnvVar(enableEventsEnvKey, "true"))
	} else {
		envProps = append(envProps, framework.CreateEnvVar(enableEventsEnvKey, "false"))
	}
	return envProps
}

func getKafkaAppProps(kafkaInstance *kafkabetav1.Kafka) map[string]string {
	kafkaURI := infrastructure.ResolveKafkaServerURI(kafkaInstance)
	appProps := map[string]string{}
	if len(kafkaURI) > 0 {
		appProps[springBootstrapAppProp] = kafkaURI
		appProps[quarkusBootstrapAppProp] = kafkaURI
	}
	return appProps
}
