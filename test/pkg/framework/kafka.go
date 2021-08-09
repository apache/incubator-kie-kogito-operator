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

package framework

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	infrastructure "github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/infrastructure/kafka/v1beta2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// DeployKafkaInstance deploys an instance of Kafka
func DeployKafkaInstance(namespace string, kafka *v1beta2.Kafka) error {
	GetLogger(namespace).Info("Creating Kafka instance %s.", "name", kafka.Name)

	if err := kubernetes.ResourceC(kubeClient).Create(kafka); err != nil {
		return fmt.Errorf("Error while creating Kafka: %v ", err)
	}

	return nil
}

// DeployKafkaTopic deploys a Kafka topic
func DeployKafkaTopic(namespace, kafkaTopicName, kafkaInstanceName string) error {
	GetLogger(namespace).Info("Creating Kafka", "topic", kafkaTopicName, "instanceName", kafkaInstanceName)

	kafkaTopic := &v1beta2.KafkaTopic{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      kafkaTopicName,
			Labels:    map[string]string{"strimzi.io/cluster": kafkaInstanceName},
		},
		Spec: v1beta2.KafkaTopicSpec{
			Replicas:   1,
			Partitions: 1,
		},
	}

	if err := kubernetes.ResourceC(kubeClient).Create(kafkaTopic); err != nil {
		return fmt.Errorf("Error while creating Kafka Topic: %v ", err)
	}

	return nil
}

// ScaleKafkaInstanceDown scales a Kafka instance down by killing its pod temporarily
func ScaleKafkaInstanceDown(namespace, kafkaInstanceName string) error {
	GetLogger(namespace).Info("Scaling Kafka Instance down", "instance name", kafkaInstanceName)
	pods, err := GetPodsWithLabels(namespace, map[string]string{"strimzi.io/name": kafkaInstanceName + "-kafka"})
	if err != nil {
		return err
	} else if len(pods.Items) != 1 {
		return fmt.Errorf("Kafka instance should have just one kafka pod running")
	}
	if err = DeleteObject(&pods.Items[0]); err != nil {
		return fmt.Errorf("Error scaling Kafka instance down by deleting a kafka pod. The nested error is: %v", err)
	}

	return nil
}

func WaitForMessagesOnTopic(namespace, kafkaInstanceName, topic string, numberOfMsg int, timeoutInMin int) error {
	return WaitForOnOpenshift(namespace, fmt.Sprintf("%d message available on topic %s withing %d minutes", numberOfMsg, topic, timeoutInMin), timeoutInMin,
		func() (bool, error) {
			messages, err := GetMessagesOnTopic(namespace, kafkaInstanceName, topic)
			if err != nil {
				return false, err
			}
			GetLogger(namespace).Info(fmt.Sprintf("Got %d messages", len(messages)))
			return len(messages) >= numberOfMsg, nil
		})
}

func GetMessagesOnTopic(namespace, kafkaInstanceName, topic string) ([]string, error) {
	GetLogger(namespace).Info("GetMessagesOnTopic")
	kafkaInstance, err := GetKafkaInstance(namespace, kafkaInstanceName)
	if err != nil {
		return nil, err
	}
	GetLogger(namespace).Info("Got kafka instance", "instance", kafkaInstance.Name)
	bootstrapServer := infrastructure.ResolveKafkaServerURI(kafkaInstance)
	if len(bootstrapServer) <= 0 {
		GetLogger(namespace).Debug("Not able resolve URI for given kafka instance")
		return nil, fmt.Errorf("not able resolve URI for given kafka instance %s", kafkaInstance.Name)
	}
	GetLogger(namespace).Info("Got bootstrapServer", "server", bootstrapServer)

	args := []string{"run", "-ti", "kafka-consumer", "--restart=Never"}
	args = append(args, fmt.Sprintf("--image=%s", "quay.io/strimzi/kafka:0.24.0-kafka-2.8.0")) // TODO set as var for testing
	args = append(args, "-n", namespace)
	args = append(args, "--")
	args = append(args, "bin/kafka-console-consumer.sh")
	args = append(args, "--bootstrap-server", bootstrapServer)
	args = append(args, "--topic", topic)
	args = append(args, "--from-beginning")
	args = append(args, "--timeout-ms", "10000")

	_, err = CreateCommand("kubectl", args...).WithLoggerContext(namespace).Execute()
	if err != nil {
		return nil, err
	}

	var output string
	output, err = CreateCommand("kubectl", "logs", "kafka-consumer", "-n", namespace).WithLoggerContext(namespace).Execute()
	if err != nil {
		return nil, err
	}
	GetLogger(namespace).Info("Got output", "output", output)
	lines := strings.Split(output, "\r\n")

	var messages []string
	var result map[string]interface{}
	for _, line := range lines {
		err = json.Unmarshal([]byte(line), &result)
		if err == nil {
			messages = append(messages, line)
		}
	}
	GetLogger(namespace).Info(fmt.Sprintf("Got %d messages", len(messages)))

	_, err = CreateCommand("kubectl", "delete", "pod", "kafka-consumer").WithLoggerContext(namespace).Execute()
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func GetKafkaInstance(namespace, kafkaInstanceName string) (*v1beta2.Kafka, error) {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      kafkaInstanceName,
	}
	kafkaInstance := &v1beta2.Kafka{}
	if exists, err := kubernetes.ResourceC(kubeClient).FetchWithKey(key, kafkaInstance); err != nil {
		GetLogger(namespace).Error(err, "Error occurs while fetching kogito kafka instance")
		return nil, err
	} else if !exists {
		GetLogger(namespace).Error(err, "kafka instance does not exist")
		return nil, nil
	} else {
		GetLogger(namespace).Debug("kafka instance found")
		return kafkaInstance, nil
	}
}
