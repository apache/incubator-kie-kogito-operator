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
	"reflect"
	"time"

	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/record"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	Deploy() (reconcileAfter time.Duration, err error)
}

// NewServiceDeployer creates a new ServiceDeployer to handle a custom Kogito Service instance to be handled by Operator SDK controller.
func NewServiceDeployer(definition ServiceDefinition, serviceType v1alpha1.KogitoService, cli *client.Client, scheme *runtime.Scheme) ServiceDeployer {
	if len(definition.Request.NamespacedName.Namespace) == 0 && len(definition.Request.NamespacedName.Name) == 0 {
		panic("No Request provided for the Service Deployer")
	}
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

type serviceDeployer struct {
	definition ServiceDefinition
	instance   v1alpha1.KogitoService
	client     *client.Client
	scheme     *runtime.Scheme
	recorder   record.EventRecorder
}

func (s *serviceDeployer) getNamespace() string { return s.definition.Request.Namespace }

func (s *serviceDeployer) Deploy() (time.Duration, error) {
	if s.instance.GetSpec().GetReplicas() == nil {
		s.instance.GetSpec().SetReplicas(defaultReplicas)
	}
	if len(s.definition.DefaultImageName) == 0 {
		s.definition.DefaultImageName = s.definition.Request.Name
	}

	var err error

	// always updateStatus its status
	defer s.handleStatusUpdate(s.instance, &err)

	// we need to take ownership of the custom configmap provided
	if len(s.instance.GetSpec().GetPropertiesConfigMap()) > 0 {
		reconcileAfter, err := s.takeCustomConfigMapOwnership()
		if err != nil || reconcileAfter > 0 {
			return reconcileAfter, err
		}
	}

	// we need to take ownership of the provided KogitoInfra instances
	if len(s.instance.GetSpec().GetInfra()) > 0 {
		err = s.takeKogitoInfraOwnership()
		if err != nil {
			return s.getReconcileResultFor(err)
		}
	}

	if err = s.checkInfraDependencies(); err != nil {
		return s.getReconcileResultFor(err)
	}

	// create our resources
	requestedResources, err := s.createRequiredResources()
	if err != nil {
		return s.getReconcileResultFor(err)
	}

	// get the deployed ones
	deployedResources, err := s.getDeployedResources()
	if err != nil {
		return s.getReconcileResultFor(err)
	}

	// compare required and deployed, in case of any differences, we should create updateStatus or delete the k8s resources
	comparator := s.getComparator()
	deltas := comparator.Compare(deployedResources, requestedResources)
	for resourceType, delta := range deltas {
		if !delta.HasChanges() {
			continue
		}
		log.Infof("Will create %d, update %d, and delete %d instances of %v", len(delta.Added), len(delta.Updated), len(delta.Removed), resourceType)

		if _, err = kubernetes.ResourceC(s.client).CreateResources(delta.Added); err != nil {
			return s.getReconcileResultFor(err)
		}
		s.generateEventForDeltaResources("Created", resourceType, delta.Added)

		if _, err = kubernetes.ResourceC(s.client).UpdateResources(deployedResources[resourceType], delta.Updated); err != nil {
			return s.getReconcileResultFor(err)
		}
		s.generateEventForDeltaResources("Updated", resourceType, delta.Updated)

		if _, err = kubernetes.ResourceC(s.client).DeleteResources(delta.Removed); err != nil {
			return s.getReconcileResultFor(err)
		}
		s.generateEventForDeltaResources("Removed", resourceType, delta.Removed)
	}

	err = s.configureMonitoring()
	if err != nil {
		return s.getReconcileResultFor(err)
	}

	err = s.configureMessaging()

	return s.getReconcileResultFor(err)
}

func (s *serviceDeployer) generateEventForDeltaResources(eventReason string, resourceType reflect.Type, addedResources []resource.KubernetesResource) {
	for _, newResource := range addedResources {
		s.recorder.Eventf(s.client, s.instance, v1.EventTypeNormal, eventReason, "%s %s: %s", eventReason, resourceType.Name(), newResource.GetName())
	}
}

func (s *serviceDeployer) takeCustomConfigMapOwnership() (requeueAfter time.Duration, err error) {
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
	return time.Second * 15, nil
}

func (s *serviceDeployer) takeKogitoInfraOwnership() (err error) {
	for _, infraName := range s.instance.GetSpec().GetInfra() {
		kogitoInfra, err := infrastructure.MustFetchKogitoInfraInstance(s.client, infraName, s.getNamespace())
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

// checkInfraDependencies verifies if every KogitoInfra resource have an ok status.
func (s *serviceDeployer) checkInfraDependencies() error {
	kogitoInfraReferences := s.instance.GetSpec().GetInfra()
	log.Debugf("Going to fetch kogito infra properties for given references : %s", kogitoInfraReferences)
	for _, infraName := range kogitoInfraReferences {
		infra, err := infrastructure.MustFetchKogitoInfraInstance(s.client, infraName, s.instance.GetNamespace())
		if err != nil {
			return err
		}
		if infra.Status.Condition.Type == v1alpha1.FailureInfraConditionType {
			return errorForInfraNotReady(s.instance, infra)
		}
	}
	return nil
}

func (s *serviceDeployer) configureMessaging() error {
	if err := handleMessagingResources(s.client, s.scheme, s.definition, s.instance); err != nil {
		return errorForMessaging(err)
	}
	return nil
}

func (s *serviceDeployer) configureMonitoring() error {
	if err := configurePrometheus(s.client, s.instance, s.scheme); err != nil {
		return err
	}

	if err := configureGrafanaDashboards(s.client, s.instance, s.scheme, s.getNamespace()); err != nil {
		log.Warnf("Could not deploy grafana dashboards due to %v", err)
		return err
	}

	return nil
}

func (s *serviceDeployer) getReconcileResultFor(err error) (time.Duration, error) {
	// reconciliation always happens if we return an error
	if reasonForError(err) == v1alpha1.ServiceReconciliationFailure {
		return 0, err
	}
	reconcileAfter := reconciliationIntervalForError(err)
	return reconcileAfter, nil
}
