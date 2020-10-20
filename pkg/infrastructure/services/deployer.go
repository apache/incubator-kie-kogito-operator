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
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/record"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"

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
	// KafkaTopics is a collection of Kafka Topics to be created within the service
	KafkaTopics []string
	// HealthCheckProbe is the probe that needs to be configured in the service. Defaults to TCPHealthCheckProbe
	HealthCheckProbe HealthCheckProbeType
	// CustomService indicates that the service can be built within the cluster
	// A custom service means that could be built by a third party, not being provided by the Kogito Team Services catalog (such as Data Index, Management Console and etc.).
	CustomService bool
	// extraManagedObjectLists is a holder for the OnObjectsCreate return function
	extraManagedObjectLists []runtime.Object
}

const (
	defaultReplicas = int32(1)
)

// ServiceDeployer is the API to handle a Kogito Service deployment by Operator SDK controllers
type ServiceDeployer interface {
	// Deploy deploys the Kogito Service in the Kubernetes cluster according to a given ServiceDefinition
	Deploy() (requeue bool, err error)
}

// NewServiceDeployer creates a new ServiceDeployer to handle a custom Kogito Service instance to be handled by Operator SDK controller.
func NewServiceDeployer(definition ServiceDefinition, serviceType v1alpha1.KogitoService, cli *client.Client, scheme *runtime.Scheme) ServiceDeployer {
	builderCheck(definition)
	return &serviceDeployer{
		definition: definition,
		instance:   serviceType,
		client:     cli,
		scheme:     scheme,
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
	definition ServiceDefinition
	instance   v1alpha1.KogitoService
	client     *client.Client
	scheme     *runtime.Scheme
	recorder   record.EventRecorder
}

func (s *serviceDeployer) getNamespace() string { return s.definition.Request.Namespace }

func (s *serviceDeployer) Deploy() (requeue bool, err error) {
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
		requeue, err = s.takeCustomConfigMapOwnership()
		if err != nil {
			return
		} else if requeue {
			return
		}
	}

	// we need to take ownership of the provided KogitoInfra instances
	if len(s.instance.GetSpec().GetInfra()) > 0 {
		err = s.takeKogitoInfraOwnership()
		if err != nil {
			return
		}
	}

	// create our resources
	requestedResources, err := s.createRequiredResources()
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
	err = configurePrometheus(s.client, s.instance, s.scheme)
	return
}

func (s *serviceDeployer) generateEventForDeltaResources(eventReason string, resourceType reflect.Type, addedResources []resource.KubernetesResource) {
	for _, newResource := range addedResources {
		s.recorder.Eventf(s.client, s.instance, v1.EventTypeNormal, eventReason, "%s %s: %s", eventReason, resourceType.Name(), newResource.GetName())
	}
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

func (s *serviceDeployer) takeCustomConfigMapOwnership() (requeue bool, err error) {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: s.instance.GetSpec().GetPropertiesConfigMap(), Namespace: s.getNamespace()},
	}
	exists, err := kubernetes.ResourceC(s.client).Fetch(cm)
	if err != nil {
		return
	}
	if !exists {
		s.recorder.Eventf(s.client, s.instance, v1.EventTypeWarning, "NotExists", "ConfigMap %s does not exist a new one will be created", s.instance.GetSpec().GetPropertiesConfigMap())
		return
	}
	if framework.IsOwner(cm, s.instance) {
		return
	}
	if err = framework.AddOwnerReference(s.instance, s.scheme, cm); err != nil {
		return
	}
	if err = kubernetes.ResourceC(s.client).Update(cm); err != nil {
		return
	}
	return true, nil
}

func (s *serviceDeployer) takeKogitoInfraOwnership() (err error) {
	for _, infraName := range s.instance.GetSpec().GetInfra() {
		kogitoInfra, err := infrastructure.FetchKogitoInfraInstance(s.client, infraName, s.getNamespace())
		if err != nil {
			return err
		}
		if framework.IsOwner(kogitoInfra, s.instance) {
			continue
		}
		if err = framework.AddOwnerReference(s.instance, s.scheme, kogitoInfra); err != nil {
			return err
		}
		if err = kubernetes.ResourceC(s.client).Update(kogitoInfra); err != nil {
			return err
		}
	}
	return nil
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
