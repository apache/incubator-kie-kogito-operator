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

package kogitoinfra

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sort"
	"strings"
	"time"
)

const (
	enableEventsEnvKey = "ENABLE_EVENTS"

	// springKafkaBootstrapAppProp spring boot application property for setting kafka server
	springKafkaBootstrapAppProp = "spring.kafka.bootstrap-servers"

	kafkaDefaultReplicas = 1
)

func getKafkaEnvVars(kafkaInstance *kafkabetav1.Kafka) ([]corev1.EnvVar, error) {
	kafkaURI, err := infrastructure.ResolveKafkaServerURI(kafkaInstance)
	if err != nil {
		return nil, err
	}
	var envProps []corev1.EnvVar
	if len(kafkaURI) > 0 {
		envProps = append(envProps, framework.CreateEnvVar(enableEventsEnvKey, "true"))
	} else {
		envProps = append(envProps, framework.CreateEnvVar(enableEventsEnvKey, "false"))
	}
	return envProps, nil
}

func getKafkaAppProps(kafkaInstance *kafkabetav1.Kafka) (map[string]string, error) {
	kafkaURI, err := infrastructure.ResolveKafkaServerURI(kafkaInstance)
	if err != nil {
		return nil, err
	}
	appProps := map[string]string{}
	if len(kafkaURI) > 0 {
		appProps[springKafkaBootstrapAppProp] = kafkaURI
		appProps[services.QuarkusKafkaBootstrapAppProp] = kafkaURI
	}
	return appProps, nil
}

// kafkaInfraResource implementation of KogitoInfraResource
type kafkaInfraResource struct {
}

// getKafkaWatchedObjects provide list of object that needs to be watched to maintain Kafka kogitoInfra resource
func getKafkaWatchedObjects() []framework.WatchedObjects {
	return []framework.WatchedObjects{
		{
			GroupVersion: kafkabetav1.SchemeGroupVersion,
			AddToScheme:  kafkabetav1.SchemeBuilder.AddToScheme,
			Objects:      []runtime.Object{&kafkabetav1.Kafka{}},
		},
	}
}

// Reconcile reconcile Kogito infra object
func (k *kafkaInfraResource) Reconcile(client *client.Client, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) (requeue bool, resultErr error) {
	var kafkaInstance *kafkabetav1.Kafka

	// Verify kafka
	if !infrastructure.IsStrimziAvailable(client) {
		return false, errorForResourceAPINotFound(&instance.Spec.Resource)
	}

	if len(instance.Spec.Resource.Name) > 0 {
		log.Debugf("Custom kafka instance reference is provided")
		namespace := instance.Spec.Resource.Namespace
		if len(namespace) == 0 {
			namespace = instance.Namespace
			log.Debugf("Namespace is not provided for custom resource, taking instance namespace(%s) as default", namespace)
		}
		if kafkaInstance, resultErr = loadDeployedKafkaInstance(client, instance.Spec.Resource.Name, namespace); resultErr != nil {
			return false, resultErr
		} else if kafkaInstance == nil {
			return false,
				errorForResourceNotFound("Kafka", instance.Spec.Resource.Name, namespace)
		}
	} else {
		// create/refer kogito-kafka instance
		log.Debugf("Custom kafka instance reference is not provided")

		// check whether kafka instance exist
		kafkaInstance, resultErr = loadDeployedKafkaInstance(client, infrastructure.KafkaInstanceName, instance.Namespace)
		if resultErr != nil {
			return false, resultErr
		}

		if kafkaInstance == nil {
			// if not exist then create new Kafka instance. Strimzi operator creates Kafka instance, secret & service resource
			_, resultErr = createNewKafkaInstance(client, infrastructure.KafkaInstanceName, instance.Namespace, instance, scheme)
			if resultErr != nil {
				return false, resultErr
			}
			return true, nil
		}
	}
	kafkaStatus := getLatestKafkaCondition(kafkaInstance)
	if kafkaStatus == nil || kafkaStatus.Type != kafkabetav1.KafkaConditionTypeReady {
		return false, errorForResourceNotReadyError(fmt.Errorf("kafka instance %s not ready yet. Waiting for Condition status Ready", kafkaInstance.Name))
	}
	if resultErr = updateKafkaAppPropsInStatus(kafkaInstance, instance); resultErr != nil {
		return true, resultErr
	}
	if resultErr = updateKafkaEnvVarsInStatus(kafkaInstance, instance); resultErr != nil {
		return true, resultErr
	}
	return false, nil
}

func updateKafkaAppPropsInStatus(kafkaInstance *kafkabetav1.Kafka, instance *v1alpha1.KogitoInfra) error {
	appProps, err := getKafkaAppProps(kafkaInstance)
	if err != nil {
		return errorForResourceNotReadyError(err)
	}
	instance.Status.AppProps = appProps
	return nil
}

func updateKafkaEnvVarsInStatus(kafkaInstance *kafkabetav1.Kafka, instance *v1alpha1.KogitoInfra) error {
	envVars, err := getKafkaEnvVars(kafkaInstance)
	if err != nil {
		return errorForResourceNotReadyError(err)
	}
	instance.Status.Env = envVars
	return nil
}

func loadDeployedKafkaInstance(cli *client.Client, name string, namespace string) (*kafkabetav1.Kafka, error) {
	log.Debug("fetching deployed kogito kafka instance")
	kafkaInstance := &kafkabetav1.Kafka{}
	if exists, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, kafkaInstance); err != nil {
		log.Error("Error occurs while fetching kogito kafka instance")
		return nil, err
	} else if !exists {
		log.Debug("Kogito kafka instance does not exist")
		return nil, nil
	} else {
		log.Debug("Kogito kafka instance found")
		return kafkaInstance, nil
	}
}

func createNewKafkaInstance(cli *client.Client, name, namespace string, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) (*kafkabetav1.Kafka, error) {
	log.Debug("Going to create kogito Kafka instance")
	kafkaInstance := infrastructure.GetKafkaDefaultResource(name, namespace, kafkaDefaultReplicas)
	if err := framework.SetOwner(instance, scheme, kafkaInstance); err != nil {
		return nil, err
	}
	if err := kubernetes.ResourceC(cli).Create(kafkaInstance); err != nil {
		log.Error("Error occurs while creating kogito Kafka instance")
		return nil, err
	}
	log.Debug("Kogito Kafka instance created successfully")
	return kafkaInstance, nil
}

func getLatestKafkaCondition(kafka *kafkabetav1.Kafka) *kafkabetav1.KafkaCondition {
	if len(kafka.Status.Conditions) == 0 {
		return nil
	}
	sort.Slice(kafka.Status.Conditions, func(i, j int) bool {
		t1, parsed := mustParseKafkaTransition(kafka.Status.Conditions[i].LastTransitionTime)
		if !parsed {
			return false
		}
		t2, parsed := mustParseKafkaTransition(kafka.Status.Conditions[j].LastTransitionTime)
		if !parsed {
			return false
		}
		return t1.Before(*t2)
	})
	return &kafka.Status.Conditions[len(kafka.Status.Conditions)-1]
}

func mustParseKafkaTransition(transitionTime string) (*time.Time, bool) {
	// TODO: open an issue on Strimzi to handle this! (this is the UTC being set to GMT+0)
	zoneIndex := strings.LastIndex(transitionTime, "+")
	if zoneIndex > -1 {
		transitionTime = string([]rune(transitionTime)[0:zoneIndex])
	}
	if !strings.Contains(transitionTime, "Z") {
		transitionTime = transitionTime + "Z"
	}
	parsedTime, err := time.Parse(kafkabetav1.KafkaLastTransitionTimeLayout, transitionTime)
	if err != nil {
		log.Errorf("Impossible to parse Kafka time condition: %s", transitionTime)
		return nil, false
	}
	return &parsedTime, true
}
