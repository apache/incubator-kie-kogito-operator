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

package converter

import (
	"github.com/kiegroup/kogito-cloud-operator/api/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
)

// FromKafkaFlagsToKafkaMeta converts given KafkaFlags into KafkaConnectionProperties
func FromKafkaFlagsToKafkaMeta(kafkaFlags *flag.KafkaFlags, enableEvents bool) v1alpha1.KafkaMeta {
	kafkaMeta := v1alpha1.KafkaMeta{}
	// configure Kafka connection properties only if enableEvents flag is true
	if enableEvents {
		kafkaMeta.KafkaProperties = fromKafkaFlagsToKafkaProperties(kafkaFlags)
	}
	return kafkaMeta
}

func fromKafkaFlagsToKafkaProperties(kafkaFlags *flag.KafkaFlags) v1alpha1.KafkaConnectionProperties {
	log := context.GetDefaultLogger()
	kafkaProperties := v1alpha1.KafkaConnectionProperties{}
	// If User provided Kafka ExternalURI or instance details then configure connection properties to user define values else
	// set UseKogitoInfra to true so Kafka will be automatically deployed via Strimzi Operator
	if len(kafkaFlags.ExternalURI) > 0 || len(kafkaFlags.Instance) > 0 {
		initializeUserDefineKafkaProperties(&kafkaProperties, kafkaFlags)
	} else {
		kafkaProperties.UseKogitoInfra = true
		log.Info("No Kafka information has been given. A Kafka instance will be automatically deployed via Strimzi Operator in the namespace. Kafka Topics will be created accordingly if they don't exist already")
	}
	return kafkaProperties
}

// initializeUserDefineKafkaProperties set Kafka connection properties to user define values
func initializeUserDefineKafkaProperties(kafkaProperties *v1alpha1.KafkaConnectionProperties, kafkaFlags *flag.KafkaFlags) {
	log := context.GetDefaultLogger()
	log.Info("kafka-instance/ExternalURI informed. Kafka will NOT be provisioned for you. Make sure Kafka instance is properly deployed in the project. If the Kafka instance is found, Kafka Topics will be created accordingly if they don't exist already")
	kafkaProperties.ExternalURI = kafkaFlags.ExternalURI
	kafkaProperties.Instance = kafkaFlags.Instance
	kafkaProperties.UseKogitoInfra = false
}
