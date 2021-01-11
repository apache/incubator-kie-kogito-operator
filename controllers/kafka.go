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

package controllers

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sort"
	"strings"
	"time"

	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/api/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
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

func getKafkaRuntimeAppProps(kafkaInstance *kafkabetav1.Kafka, runtime v1beta1.RuntimeType) (map[string]string, error) {
	kafkaURI, err := infrastructure.ResolveKafkaServerURI(kafkaInstance)
	if err != nil {
		return nil, err
	}
	appProps := map[string]string{}
	if len(kafkaURI) > 0 {
		if runtime == v1beta1.QuarkusRuntimeType {
			appProps[services.QuarkusKafkaBootstrapAppProp] = kafkaURI
		} else if runtime == v1beta1.SpringBootRuntimeType {
			appProps[springKafkaBootstrapAppProp] = kafkaURI
		}
	}
	return appProps, nil
}

func initkafkaInfraReconciler(context targetContext) *kafkaInfraReconciler {
	log := logger.GetLogger("kafka")
	return &kafkaInfraReconciler{
		targetContext: context,
		log:           log,
	}
}

func getKafkaRuntimeProps(kafkaInstance *kafkabetav1.Kafka, runtime v1beta1.RuntimeType) (v1beta1.RuntimeProperties, error) {
	runtimeProps := v1beta1.RuntimeProperties{}
	appProps, err := getKafkaRuntimeAppProps(kafkaInstance, runtime)
	if err != nil {
		return runtimeProps, err
	}
	runtimeProps.AppProps = appProps

	envVars, err := getKafkaEnvVars(kafkaInstance)
	if err != nil {
		return runtimeProps, err
	}
	runtimeProps.Env = envVars

	return runtimeProps, nil
}

// kafkaInfraReconciler implementation of KogitoInfraResource
type kafkaInfraReconciler struct {
	targetContext
	log logger.Logger
}

// getKafkaWatchedObjects provide list of object that needs to be watched to maintain Kafka kogitoInfra resource
func appendKafkaWatchedObjects(b *builder.Builder) *builder.Builder {
	return b
}

// Reconcile reconcile Kogito infra object
func (k *kafkaInfraReconciler) Reconcile() (requeue bool, resultErr error) {
	var kafkaInstance *kafkabetav1.Kafka

	// Verify kafka
	if !infrastructure.IsStrimziAvailable(k.client) {
		return false, errorForResourceAPINotFound(&k.instance.Spec.Resource)
	}

	if len(k.instance.Spec.Resource.Name) > 0 {
		k.log.Debug("Custom kafka instance reference is provided")
		namespace := k.instance.Spec.Resource.Namespace
		if len(namespace) == 0 {
			namespace = k.instance.Namespace
			k.log.Debug("Namespace is not provided for custom resource, taking", "Namespace", namespace)
		}
		if kafkaInstance, resultErr = k.loadDeployedKafkaInstance(k.instance.Spec.Resource.Name, namespace); resultErr != nil {
			return false, resultErr
		} else if kafkaInstance == nil {
			return false,
				errorForResourceNotFound("Kafka", k.instance.Spec.Resource.Name, namespace)
		}
	} else {
		// create/refer kogito-kafka instance
		k.log.Debug("Custom kafka instance reference is not provided")

		// check whether kafka instance exist
		kafkaInstance, resultErr = k.loadDeployedKafkaInstance(infrastructure.KafkaInstanceName, k.instance.Namespace)
		if resultErr != nil {
			return false, resultErr
		}

		if kafkaInstance == nil {
			// if not exist then create new Kafka instance. Strimzi operator creates Kafka instance, secret & service resource
			_, resultErr = k.createNewKafkaInstance(infrastructure.KafkaInstanceName, k.instance.Namespace)
			if resultErr != nil {
				return false, resultErr
			}
			return true, nil
		}
	}
	kafkaStatus := k.getLatestKafkaCondition(kafkaInstance)
	if kafkaStatus == nil || kafkaStatus.Type != kafkabetav1.KafkaConditionTypeReady {
		return false, errorForResourceNotReadyError(fmt.Errorf("kafka instance %s not ready yet. Waiting for Condition status Ready", kafkaInstance.Name))
	}
	if resultErr = k.updateKafkaRuntimePropsInStatus(kafkaInstance, v1beta1.QuarkusRuntimeType); resultErr != nil {
		return true, resultErr
	}
	if resultErr = k.updateKafkaRuntimePropsInStatus(kafkaInstance, v1beta1.SpringBootRuntimeType); resultErr != nil {
		return true, resultErr
	}
	return false, nil
}

func (k *kafkaInfraReconciler) updateKafkaRuntimePropsInStatus(kafkaInstance *kafkabetav1.Kafka, runtime v1beta1.RuntimeType) error {
	k.log.Debug("going to Update Kafka runtime properties in kogito infra instance status", "runtime", runtime)
	runtimeProps, err := getKafkaRuntimeProps(kafkaInstance, runtime)
	if err != nil {
		return errorForResourceNotReadyError(err)
	}
	setRuntimeProperties(k.instance, runtime, runtimeProps)
	k.log.Debug("Following Kafka runtime properties are set in infra status:", "runtime", runtime, "properties", runtimeProps)
	return nil
}

func (k *kafkaInfraReconciler) loadDeployedKafkaInstance(name, namespace string) (*kafkabetav1.Kafka, error) {
	k.log.Debug("fetching deployed kogito kafka instance")
	kafkaInstance := &kafkabetav1.Kafka{}
	if exists, err := kubernetes.ResourceC(k.client).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, kafkaInstance); err != nil {
		k.log.Error(err, "Error occurs while fetching kogito kafka instance")
		return nil, err
	} else if !exists {
		k.log.Debug("Kogito kafka instance does not exist")
		return nil, nil
	} else {
		k.log.Debug("Kogito kafka instance found")
		return kafkaInstance, nil
	}
}

func (k *kafkaInfraReconciler) createNewKafkaInstance(name, namespace string) (*kafkabetav1.Kafka, error) {
	k.log.Debug("Going to create kogito Kafka instance")
	kafkaInstance := infrastructure.GetKafkaDefaultResource(name, namespace, kafkaDefaultReplicas)
	if err := framework.SetOwner(k.instance, k.scheme, kafkaInstance); err != nil {
		return nil, err
	}
	if err := kubernetes.ResourceC(k.client).Create(kafkaInstance); err != nil {
		k.log.Error(err, "Error occurs while creating kogito Kafka instance")
		return nil, err
	}
	k.log.Debug("Kogito Kafka instance created successfully")
	return kafkaInstance, nil
}

func (k *kafkaInfraReconciler) getLatestKafkaCondition(kafka *kafkabetav1.Kafka) *kafkabetav1.KafkaCondition {
	if len(kafka.Status.Conditions) == 0 {
		return nil
	}
	sort.Slice(kafka.Status.Conditions, func(i, j int) bool {
		t1, parsed := k.mustParseKafkaTransition(kafka.Status.Conditions[i].LastTransitionTime)
		if !parsed {
			return false
		}
		t2, parsed := k.mustParseKafkaTransition(kafka.Status.Conditions[j].LastTransitionTime)
		if !parsed {
			return false
		}
		return t1.Before(*t2)
	})
	return &kafka.Status.Conditions[len(kafka.Status.Conditions)-1]
}

func (k *kafkaInfraReconciler) mustParseKafkaTransition(transitionTime string) (*time.Time, bool) {
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
		k.log.Error(err, "Impossible to parse", "Kafka time condition", transitionTime)
		return nil, false
	}
	return &parsedTime, true
}
