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
	"reflect"
	"time"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/RHsyseng/operator-utils/pkg/resource/write"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controller "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var log = logger.GetLogger("services_definition")

const (
	reconciliationPeriodAfterSingletonError = time.Minute
)

// ServiceDefinition defines the structure for a Kogito Service
type ServiceDefinition struct {
	// DefaultImageName is the name of the default image distributed for Kogito, e.g. kogito-jobs-service, kogito-data-index and so on
	// can be empty, in this case Request.Name will be used as image name
	DefaultImageName string
	// DefaultImageTag is the default image tag to use for this service. If left empty, will use the minor version of the operator, e.g. 0.11
	DefaultImageTag string
	// Request made for the service
	Request controller.Request
	// OnDeploymentCreate applies custom deployment configuration in the required Deployment resource
	OnDeploymentCreate func(deployment *appsv1.Deployment, kogitoService v1alpha1.KogitoService) error
	// OnObjectsCreate applies custom object creation in the service deployment logic.
	// E.g. if you need an additional Kubernetes resource, just create your own map that the API will append to its managed resources.
	// The "objectLists" array is the List object reference of the types created.
	// For example: if a ConfigMap is created, then ConfigMapList empty reference should be added to this list
	OnObjectsCreate func(kogitoService v1alpha1.KogitoService) (resources map[reflect.Type][]resource.KubernetesResource, objectLists []runtime.Object, err error)
	// OnGetComparators is called during the deployment phase to compare the deployed resources against the created ones
	// Use this hook to add your comparators to override a specific comparator or to add your own if you have created extra objects via OnObjectsCreate
	// Use framework.NewComparatorBuilder() to build your own
	OnGetComparators func(comparator compare.ResourceComparator)
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
	// extraManagedObjectLists is a holder for the OnObjectsCreate return function
	extraManagedObjectLists []runtime.Object
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
	defaultReplicas             = int32(1)
	serviceDoesNotExistsMessage = "Kogito Service '%s' does not exists, aborting deployment"
)

// ServiceDeployer is the API to handle a Kogito Service deployment by Operator SDK controllers
type ServiceDeployer interface {
	// Deploy deploys the Kogito Service in the Kubernetes cluster according to a given ServiceDefinition
	Deploy() (reconcileAfter time.Duration, err error)
}

// NewSingletonServiceDeployer creates a new ServiceDeployer to handle Singleton Kogito Services instances and to be handled by Operator SDK controller
func NewSingletonServiceDeployer(definition ServiceDefinition, serviceList v1alpha1.KogitoServiceList, cli *client.Client, scheme *runtime.Scheme) ServiceDeployer {
	builderCheck(definition)
	return &serviceDeployer{definition: definition, instanceList: serviceList, client: cli, scheme: scheme, singleton: true}
}

// NewServiceDeployer creates a new ServiceDeployer to handle a Kogito Service instance and to be handled by Operator SDK controller
func NewServiceDeployer(definition ServiceDefinition, serviceType v1alpha1.KogitoService, cli *client.Client, scheme *runtime.Scheme) ServiceDeployer {
	builderCheck(definition)
	return &serviceDeployer{definition: definition, instance: serviceType, client: cli, scheme: scheme, singleton: false}
}

func builderCheck(definition ServiceDefinition) {
	if &definition.Request == nil {
		panic("No Request provided for the Service Deployer")
	}
}

type serviceDeployer struct {
	definition   ServiceDefinition
	instanceList v1alpha1.KogitoServiceList
	instance     v1alpha1.KogitoService
	singleton    bool
	client       *client.Client
	scheme       *runtime.Scheme
}

func (s *serviceDeployer) getNamespace() string { return s.definition.Request.Namespace }

func (s *serviceDeployer) getServiceName() string { return s.definition.Request.Name }

func (s *serviceDeployer) Deploy() (reconcileAfter time.Duration, err error) {
	found, reconcileAfter, err := s.getService()
	if err != nil || !found {
		return reconcileAfter, err
	}
	if s.instance.GetSpec().GetReplicas() == nil {
		s.instance.GetSpec().SetReplicas(defaultReplicas)
	}
	if len(s.definition.DefaultImageName) == 0 {
		s.definition.DefaultImageName = s.definition.Request.Name
	}

	// always update its status
	defer s.updateStatus(s.instance, &err)

	if _, isInfinispan := s.instance.GetSpec().(v1alpha1.InfinispanAware); isInfinispan {
		log.Debugf("Kogito Service %s supports Infinispan", s.instance.GetName())
		s.definition.infinispanAware = true
	}
	if _, isKafka := s.instance.GetSpec().(v1alpha1.KafkaAware); isKafka {
		log.Debugf("Kogito Service %s supports Kafka", s.instance.GetName())
		s.definition.kafkaAware = true
	}

	// deploy Infinispan
	if s.definition.infinispanAware {
		reconcileAfter, err = s.deployInfinispan()
		if err != nil {
			return
		} else if reconcileAfter > 0 {
			return
		}
	}

	// deploy Kafka
	if s.definition.kafkaAware {
		reconcileAfter, err = s.deployKafka()
		if err != nil {
			return
		} else if reconcileAfter > 0 {
			return
		}
	}

	// create our resources
	requestedResources, reconcileAfter, err := s.createRequiredResources()
	if err != nil {
		return
	}

	// get the deployed ones
	deployedResources, err := s.getDeployedResources()
	if err != nil {
		return
	}

	// compare required and deployed, in case of any differences, we should create update or delete the k8s resources
	comparator := s.getComparator()
	deltas := comparator.Compare(deployedResources, requestedResources)
	writer := write.New(s.client.ControlCli).WithOwnerController(s.instance, s.scheme)
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

func (s *serviceDeployer) getService() (found bool, reconcileAfter time.Duration, err error) {
	reconcileAfter = 0
	if s.singleton {
		// our services must be singleton instances
		if exists, err := s.ensureSingletonService(); err != nil {
			return false, reconciliationPeriodAfterSingletonError, err
		} else if !exists {
			log.Debugf(serviceDoesNotExistsMessage, s.definition.Request.Name)
			return false, reconcileAfter, err
		}
		// we get our service
		s.instance = s.instanceList.GetItemAt(0)
	} else {
		if exists, err := kubernetes.ResourceC(s.client).FetchWithKey(s.definition.Request.NamespacedName, s.instance); err != nil {
			return false, reconcileAfter, err
		} else if !exists {
			log.Debugf(serviceDoesNotExistsMessage, s.definition.Request.Name)
			return false, reconcileAfter, nil
		}
	}
	return true, reconcileAfter, nil
}

func (s *serviceDeployer) ensureSingletonService() (exists bool, err error) {
	if err := kubernetes.ResourceC(s.client).ListWithNamespace(s.getNamespace(), s.instanceList); err != nil {
		return false, err
	}
	if s.instanceList.GetItemsCount() > 1 {
		return true, fmt.Errorf("There's more than one Kogito Service resource in the namespace %s, please delete one of them ", s.getNamespace())
	}
	return s.instanceList.GetItemsCount() > 0, nil
}

func (s *serviceDeployer) updateStatus(instance v1alpha1.KogitoService, err *error) {
	log.Infof("Updating status for Kogito Service %s", instance.GetName())
	if statusErr := s.manageStatus(s.definition.DefaultImageName, s.definition.DefaultImageTag, *err); statusErr != nil {
		// this error will return to the operator console
		err = &statusErr
	}
	log.Infof("Successfully reconciled Kogito Service %s", instance.GetName())
}

func (s *serviceDeployer) deployInfinispan() (requeueAfter time.Duration, err error) {
	requeueAfter = 0
	infinispanAware := s.instance.GetSpec().(v1alpha1.InfinispanAware)
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
		log.Debugf("Service %s requires persistence and Infinispan URL is empty, deploying Kogito Infrastructure", s.instance.GetName())
		infinispanAware.GetInfinispanProperties().UseKogitoInfra = true
	} else if !infinispanAware.GetInfinispanProperties().UseKogitoInfra {
		return
	}
	if !infrastructure.IsInfinispanAvailable(s.client) {
		log.Warnf("Looks like that the service %s requires Infinispan, but there's no Infinispan CRD in the namespace %s. Aborting installation.", s.instance.GetName(), s.instance.GetNamespace())
		return
	}
	needUpdate := false
	if needUpdate, requeueAfter, err =
		deployInfinispanWithKogitoInfra(infinispanAware, s.instance.GetNamespace(), s.client); err != nil {
		return
	} else if needUpdate {
		if err = s.update(); err != nil {
			return
		}
	}
	return
}

func (s *serviceDeployer) deployKafka() (requeueAfter time.Duration, err error) {
	requeueAfter = 0
	kafkaAware := s.instance.GetSpec().(v1alpha1.KafkaAware)
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
		log.Debugf("Service %s requires messaging and Kafka URL is empty and kafka instance is not provided, deploying Kogito Infrastructure", s.instance.GetName())
		kafkaAware.GetKafkaProperties().UseKogitoInfra = true
	} else if !kafkaAware.GetKafkaProperties().UseKogitoInfra {
		return
	}
	if !infrastructure.IsStrimziAvailable(s.client) {
		log.Warnf("Looks like that the service %s requires Kafka, but there's no Kafka CRD in the namespace %s. Aborting installation.", s.instance.GetName(), s.instance.GetNamespace())
		return
	}

	needUpdate := false
	if needUpdate, requeueAfter, err =
		infrastructure.DeployKafkaWithKogitoInfra(kafkaAware, s.instance.GetNamespace(), s.client); err != nil {
		return
	} else if needUpdate {
		if err = s.update(); err != nil {
			return
		}
	}
	return
}

func (s *serviceDeployer) update() error {
	// Sanity check since the Status CR needs a reference for the object
	if s.instance.GetStatus() != nil && s.instance.GetStatus().GetConditions() == nil {
		s.instance.GetStatus().SetConditions([]v1alpha1.Condition{})
	}
	err := kubernetes.ResourceC(s.client).Update(s.instance)
	if err != nil {
		return err
	}
	return nil
}
