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
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource/write"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controller "sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

var log = logger.GetLogger("services_definition")

const (
	reconciliationPeriodAfterSingletonError = time.Minute
)

// ServiceDefinition defines the structure for a Kogito Service
type ServiceDefinition struct {
	// DefaultImageName is the name of the default image distributed for Kogito, e.g. kogito-jobs-service, kogito-data-index and so on
	DefaultImageName string
	// Request made for the service
	Request controller.Request
	// OnDeploymentCreate applies custom deployment configuration in the required Deployment resource
	OnDeploymentCreate func(deployment *appsv1.Deployment, kogitoService v1alpha1.KogitoService) error
	// SingleReplica if set to true, avoids that the service has more than one pod replica
	SingleReplica bool
	// RequiresPersistence forces the deployer to deploy an Infinispan instance if none is provided
	RequiresPersistence bool
	// RequiresMessaging forces the deployer to deploy a Kafka instance if none is provided
	RequiresMessaging bool
	// RequiresDataIndex when set to true, the Data Index instance is queried in the given namespace and its Route injected in this service.
	// The service is not deployed until the data index service is found
	RequiresDataIndex bool
	// KafkaTopics is a collection of Kafka Topics to be created within the service
	KafkaTopics []KafkaTopicDefinition
	// HealthCheckProbe is the probe that needs to be configured in the service. Defaults to TCPHealthCheckProbe
	HealthCheckProbe HealthCheckProbeType
	// infinispanAware whether or not to handle Infinispan integration in this service (inject variables, deploy if needed, and so on)
	infinispanAware bool
	// kafkaAware whether or not to handle Kafka integration in this service (inject variables, deploy if needed, and so on)
	kafkaAware bool
}

// KafkaTopicDefinition ...
type KafkaTopicDefinition struct {
	// TopicName name of the given topic
	TopicName string
	// MessagingType is the type for the Kafka topic: INCOMING or OUTGOING
	MessagingType KafkaTopicMessagingType
}

// KafkaTopicMessagingType ...
type KafkaTopicMessagingType string

const (
	// KafkaTopicIncoming ...
	KafkaTopicIncoming KafkaTopicMessagingType = "INCOMING"
	// KafkaTopicOutgoing ...
	KafkaTopicOutgoing KafkaTopicMessagingType = "OUTGOING"
)

const (
	defaultReplicas = int32(1)
)

// ServiceDeployer is the API to handle a Kogito Service deployment by Operator SDK controllers
type ServiceDeployer interface {
	// Deploy deploys the Kogito Service in the Kubernetes cluster according to a given ServiceDefinition
	Deploy() (reconcileAfter time.Duration, err error)
}

// NewSingletonServiceDeployer creates a new ServiceDeployer to handle Singleton Kogito Services instances and to be handled by Operator SDK controller
func NewSingletonServiceDeployer(definition ServiceDefinition, serviceList v1alpha1.KogitoServiceList, cli *client.Client, scheme *runtime.Scheme) ServiceDeployer {
	if &definition.Request == nil {
		panic("No Request provided for the Service Deployer")
	}
	return &serviceDeployer{definition: definition, instanceList: serviceList, client: cli, scheme: scheme}
}

type serviceDeployer struct {
	definition   ServiceDefinition
	instanceList v1alpha1.KogitoServiceList
	client       *client.Client
	scheme       *runtime.Scheme
}

func (s *serviceDeployer) getNamespace() string { return s.definition.Request.Namespace }

func (s *serviceDeployer) getServiceName() string { return s.definition.Request.Name }

func (s *serviceDeployer) Deploy() (reconcileAfter time.Duration, err error) {
	// our services must be singleton instances
	if reconcile, exists, err := s.ensureSingletonService(); err != nil || reconcile {
		return reconciliationPeriodAfterSingletonError, err
	} else if !exists {
		log.Debugf("Kogito Service '%s' does not exists, aborting deployment", s.definition.Request.Name)
		return 0, err
	}

	// we get our service
	service := s.instanceList.GetItemAt(0)
	reconcileAfter = 0

	if service.GetSpec().GetReplicas() == nil {
		service.GetSpec().SetReplicas(defaultReplicas)
	}

	// always update its status
	defer s.updateStatus(service, &err)

	if _, isInfinispan := service.GetSpec().(v1alpha1.InfinispanAware); isInfinispan {
		log.Debugf("Kogito Service %s depends on Infinispan", service.GetName())
		s.definition.infinispanAware = true
	}
	if _, isKafka := service.GetSpec().(v1alpha1.KafkaAware); isKafka {
		log.Debugf("Kogito Service %s depends on Kafka", service.GetName())
		s.definition.kafkaAware = true
	}

	// deploy Infinispan
	if s.definition.infinispanAware {
		reconcileAfter, err = s.deployInfinispan(service)
		if err != nil {
			return
		} else if reconcileAfter > 0 {
			return
		}
	}

	// deploy Kafka
	if s.definition.kafkaAware {
		reconcileAfter, err = s.deployKafka(service)
		if err != nil {
			return
		} else if reconcileAfter > 0 {
			return
		}
	}

	// create our resources
	requestedResources, reconcileAfter, err := s.createRequiredResources(service)
	if err != nil {
		return
	}

	// get the deployed ones
	deployedResources, err := s.getDeployedResources(service)
	if err != nil {
		return
	}

	// compare required and deployed, in case of any differences, we should create update or delete the k8s resources
	comparator := s.getComparator()
	deltas := comparator.Compare(deployedResources, requestedResources)
	writer := write.New(s.client.ControlCli).WithOwnerController(service, s.scheme)
	for resourceType, delta := range deltas {
		if !delta.HasChanges() {
			continue
		}
		log.Infof("Will create %d, update %d, and delete %d instances of %v", len(delta.Added), len(delta.Updated), len(delta.Removed), resourceType)
		_, err = writer.AddResources(delta.Added)
		if err != nil {
			return
		}
		_, err = writer.UpdateResources(deployedResources[resourceType], delta.Updated)
		if err != nil {
			return
		}
		_, err = writer.RemoveResources(delta.Removed)
		if err != nil {
			return
		}
	}

	return
}

func (s *serviceDeployer) ensureSingletonService() (reconcile bool, exists bool, err error) {
	if err := kubernetes.ResourceC(s.client).ListWithNamespace(s.getNamespace(), s.instanceList); err != nil {
		return true, false, err
	}
	if s.instanceList.GetItemsCount() > 1 {
		return true, true, fmt.Errorf("There's more than one Kogito Service resource in the namespace %s, please delete one of them ", s.getNamespace())
	}
	return false, s.instanceList.GetItemsCount() > 0, nil
}

func (s *serviceDeployer) updateStatus(instance v1alpha1.KogitoService, err *error) {
	log.Infof("Updating status for Kogito Service %s", instance.GetName())
	if statusErr := s.manageStatus(instance, s.definition.DefaultImageName, *err); statusErr != nil {
		// this error will return to the operator console
		err = &statusErr
	}
	log.Infof("Successfully reconciled Kogito Service %s", instance.GetName())
}

func (s *serviceDeployer) deployInfinispan(instance v1alpha1.KogitoService) (requeueAfter time.Duration, err error) {
	requeueAfter = 0
	infinispanAware := instance.GetSpec().(v1alpha1.InfinispanAware)
	if infinispanAware.GetInfinispanProperties() == nil {
		if s.definition.RequiresPersistence {
			infinispanAware.SetInfinispanProperties(v1alpha1.InfinispanConnectionProperties{UseKogitoInfra: true})
		} else {
			return
		}
	}
	if s.definition.RequiresPersistence &&
		!infinispanAware.GetInfinispanProperties().UseKogitoInfra &&
		len(infinispanAware.GetInfinispanProperties().URI) == 0 {
		log.Debugf("Service %s requires persistence and Infinispan URL is empty, deploying Kogito Infrastructure", instance.GetName())
		infinispanAware.GetInfinispanProperties().UseKogitoInfra = true
	} else if !infinispanAware.GetInfinispanProperties().UseKogitoInfra {
		return
	}
	if !infrastructure.IsInfinispanAvailable(s.client) {
		log.Warnf("Looks like that the service %s requires Infinispan, but there's no Infinispan CRD in the namespace %s. Aborting installation.", instance.GetName(), instance.GetNamespace())
		return
	}
	needUpdate := false
	if needUpdate, requeueAfter, err =
		infrastructure.DeployInfinispanWithKogitoInfra(infinispanAware, instance.GetNamespace(), s.client); err != nil {
		return
	} else if needUpdate {
		if err = s.update(instance); err != nil {
			return
		}
	}
	return
}

func (s *serviceDeployer) deployKafka(instance v1alpha1.KogitoService) (requeueAfter time.Duration, err error) {
	requeueAfter = 0
	kafkaAware := instance.GetSpec().(v1alpha1.KafkaAware)
	if kafkaAware.GetKafkaProperties() == nil {
		if s.definition.RequiresMessaging {
			kafkaAware.SetKafkaProperties(v1alpha1.KafkaConnectionProperties{UseKogitoInfra: true})
		} else {
			return
		}
	}
	if s.definition.RequiresMessaging &&
		!kafkaAware.GetKafkaProperties().UseKogitoInfra &&
		len(kafkaAware.GetKafkaProperties().ExternalURI) == 0 &&
		len(kafkaAware.GetKafkaProperties().Instance) == 0 {
		log.Debugf("Service %s requires messaging and Kafka URL is empty and kafka instance is not provided, deploying Kogito Infrastructure", instance.GetName())
		kafkaAware.GetKafkaProperties().UseKogitoInfra = true
	} else if !kafkaAware.GetKafkaProperties().UseKogitoInfra {
		return
	}
	if !infrastructure.IsStrimziAvailable(s.client) {
		log.Warnf("Looks like that the service %s requires Kafka, but there's no Kafka CRD in the namespace %s. Aborting installation.", instance.GetName(), instance.GetNamespace())
		return
	}

	needUpdate := false
	if needUpdate, requeueAfter, err =
		infrastructure.DeployKafkaWithKogitoInfra(kafkaAware, instance.GetNamespace(), s.client); err != nil {
		return
	} else if needUpdate {
		if err = s.update(instance); err != nil {
			return
		}
	}
	return
}

func (s *serviceDeployer) update(instance v1alpha1.KogitoService) error {
	// Sanity check since the Status CR needs a reference for the object
	if instance.GetStatus() != nil && instance.GetStatus().GetConditions() == nil {
		instance.GetStatus().SetConditions([]v1alpha1.Condition{})
	}
	err := kubernetes.ResourceC(s.client).Update(instance)
	if err != nil {
		return err
	}
	return nil
}
