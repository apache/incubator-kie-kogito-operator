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

	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/record"
	v1 "k8s.io/api/core/v1"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
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
	OnDeploymentCreate func(cli *client.Client, deployment *appsv1.Deployment, kogitoService v1alpha1.KogitoService) error
	// OnObjectsCreate applies custom object creation in the service deployment logic.
	// E.g. if you need an additional Kubernetes resource, just create your own map that the API will append to its managed resources.
	// The "objectLists" array is the List object reference of the types created.
	// For example: if a ConfigMap is created, then ConfigMapList empty reference should be added to this list
	OnObjectsCreate func(cli *client.Client, kogitoService v1alpha1.KogitoService) (resources map[reflect.Type][]resource.KubernetesResource, objectLists []runtime.Object, err error)
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
	// RequiresTrusty when set to true, the Trusty instance is queried in the given namespace and its Route injected in this service.
	// The service is not deployed until the trusty service is found
	RequiresTrusty bool
	// KafkaTopics is a collection of Kafka Topics to be created within the service
	KafkaTopics []string
	// HealthCheckProbe is the probe that needs to be configured in the service. Defaults to TCPHealthCheckProbe
	HealthCheckProbe HealthCheckProbeType
	// customService indicates that the service can be built within the cluster
	customService bool
	// extraManagedObjectLists is a holder for the OnObjectsCreate return function
	extraManagedObjectLists []runtime.Object
}

const (
	defaultReplicas             = int32(1)
	serviceDoesNotExistsMessage = "Kogito Service '%s' does not exists, aborting deployment"
)

// ServiceDeployer is the API to handle a Kogito Service deployment by Operator SDK controllers
type ServiceDeployer interface {
	// Deploy deploys the Kogito Service in the Kubernetes cluster according to a given ServiceDefinition
	Deploy() (reconcileAfter time.Duration, err error)
}

// NewSingletonServiceDeployer creates a new ServiceDeployer to handle Singleton Kogito Services instances to be handled by Operator SDK controller
func NewSingletonServiceDeployer(definition ServiceDefinition, serviceList v1alpha1.KogitoServiceList, cli *client.Client, scheme *runtime.Scheme) ServiceDeployer {
	builderCheck(definition)
	return &serviceDeployer{
		definition:   definition,
		instanceList: serviceList,
		client:       cli,
		scheme:       scheme,
		singleton:    true,
		recorder:     newRecorder(scheme, definition.Request.Name),
	}
}

// NewCustomServiceDeployer creates a new ServiceDeployer to handle a custom Kogito Service instance to be handled by Operator SDK controller.
// A custom service means that could be built by a third party, not being provided by the Kogito Team Services catalog (such as Data Index, Management Console and etc.).
func NewCustomServiceDeployer(definition ServiceDefinition, serviceType v1alpha1.KogitoService, cli *client.Client, scheme *runtime.Scheme) ServiceDeployer {
	builderCheck(definition)
	definition.customService = true
	return &serviceDeployer{
		definition: definition,
		instance:   serviceType,
		client:     cli,
		scheme:     scheme,
		singleton:  false,
		recorder:   newRecorder(scheme, definition.Request.Name),
	}
}

func newRecorder(scheme *runtime.Scheme, eventSourceName string) record.EventRecorder {
	return record.NewRecorder(scheme, v1.EventSource{Component: eventSourceName, Host: record.GetHostName()})
}

func builderCheck(definition ServiceDefinition) {
	if len(definition.Request.NamespacedName.Namespace) == 0 && len(definition.Request.NamespacedName.Name) == 0 {
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
	recorder     record.EventRecorder
}

func (s *serviceDeployer) getNamespace() string { return s.definition.Request.Namespace }

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

	// we need to take ownership of the custom configmap provided
	if len(s.instance.GetSpec().GetPropertiesConfigMap()) > 0 {
		reconcileAfter, err = s.takeCustomConfigMapOwnership()
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
	} else if reconcileAfter > 0 {
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
	for resourceType, delta := range deltas {
		if !delta.HasChanges() {
			continue
		}
		log.Infof("Will create %d, update %d, and delete %d instances of %v", len(delta.Added), len(delta.Updated), len(delta.Removed), resourceType)

		if _, err = kubernetes.ResourceC(s.client).CreateResources(delta.Added); err != nil {
			return
		}
		s.generateEventForDeltaResources("Created", resourceType, delta.Added)

		if _, err = kubernetes.ResourceC(s.client).UpdateResources(deployedResources[resourceType], delta.Updated); err != nil {
			return
		}
		s.generateEventForDeltaResources("Updated", resourceType, delta.Updated)

		if _, err = kubernetes.ResourceC(s.client).DeleteResources(delta.Removed); err != nil {
			return
		}
		s.generateEventForDeltaResources("Removed", resourceType, delta.Removed)
	}
	reconcileAfter, err = configurePrometheus(s.client, s.instance, s.scheme)
	return
}

func (s *serviceDeployer) generateEventForDeltaResources(eventReason string, resourceType reflect.Type, addedResources []resource.KubernetesResource) {
	for _, newResource := range addedResources {
		s.recorder.Eventf(s.client, s.instance, v1.EventTypeNormal, eventReason, "%s %s: %s", eventReason, resourceType.Name(), newResource.GetName())
	}
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
	if statusErr := s.manageStatus(*err); statusErr != nil {
		// this error will return to the operator console
		err = &statusErr
	}
	log.Infof("Successfully reconciled Kogito Service %s", instance.GetName())
	if *err != nil {
		log.Errorf("Error while creating kogito service: %v", *err)
	}
}

func (s *serviceDeployer) takeCustomConfigMapOwnership() (time.Duration, error) {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: s.instance.GetSpec().GetPropertiesConfigMap(), Namespace: s.getNamespace()},
	}
	if exists, err := kubernetes.ResourceC(s.client).Fetch(cm); err != nil {
		return 0, err
	} else if !exists {
		s.recorder.Eventf(s.client, s.instance, v1.EventTypeWarning, "NotExists", "ConfigMap %s does not exist a new one will be created", s.instance.GetSpec().GetPropertiesConfigMap())
		return 0, nil
	}
	if framework.IsOwner(cm, s.instance) {
		return 0, nil
	}
	if err := framework.AddOwnerReference(s.instance, s.scheme, cm); err != nil {
		return 0, err
	}
	if err := kubernetes.ResourceC(s.client).Update(cm); err != nil {
		return 0, err
	}
	return time.Second * 15, nil
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
