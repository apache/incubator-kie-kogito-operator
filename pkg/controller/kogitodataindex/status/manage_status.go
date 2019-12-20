// Copyright 2019 Red Hat, Inc. and/or its affiliates
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

package status

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitodataindex/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var log = logger.GetLogger("status_kogitodataindex")

// ManageStatus will guarantee the status changes
func ManageStatus(instance *v1alpha1.KogitoDataIndex, resources *resource.KogitoDataIndexResources, resourcesUpdate bool, reconcileError bool, client *client.Client) error {
	var err error
	var exists bool
	status := v1alpha1.KogitoDataIndexStatus{}
	currentCondition := v1alpha1.DataIndexCondition{}

	var statefulSet *appsv1.StatefulSet
	if resources.StatefulSet != nil {
		statefulSet = &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      resources.StatefulSet.Name,
				Namespace: resources.StatefulSet.Namespace,
			},
		}
		if exists, err = kubernetes.ResourceC(client).Fetch(statefulSet); err != nil {
			return err
		} else if exists {
			status.DeploymentStatus = statefulSet.Status
		} else {
			statefulSet = nil
		}
	}

	var service *corev1.Service
	if resources.Service != nil {
		service = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      resources.Service.Name,
				Namespace: resources.Service.Namespace,
			},
		}
		if exists, err = kubernetes.ResourceC(client).Fetch(service); err != nil {
			return err
		} else if exists {
			status.ServiceStatus = service.Status
		} else {
			service = nil
		}
	}

	if status.DependenciesStatus, err = checkDependenciesStatus(instance, client); err != nil {
		return err
	}

	if resources.Route != nil {
		log.Debugf("Trying to get the host for the route %s", resources.Route.Name)
		if exists, status.Route, err =
			openshift.RouteC(client).GetHostFromRoute(
				types.NamespacedName{Name: resources.Route.Name, Namespace: resources.Route.Namespace}); err != nil {
			return err
		} else if exists {
			status.Route = fmt.Sprintf("http://%s", status.Route)
		}
	} else {
		log.Debugf("Route is nil, impossible to get host to set in the status", resources.Route)
	}

	status.Conditions = instance.Status.Conditions
	currentCondition = checkCurrentCondition(statefulSet, service, resourcesUpdate, reconcileError)
	lastCondition := getLastCondition(instance)
	if lastCondition == nil || (currentCondition.Condition != lastCondition.Condition || currentCondition.Message != lastCondition.Message) {
		log.Debugf("Creating new status conditions. Actual conditions: %s. Current condition: %s", instance.Status.Conditions, currentCondition)
		if &status.Conditions == nil {
			status.Conditions = []v1alpha1.DataIndexCondition{}
		}
		status.Conditions = append(status.Conditions, currentCondition)
	}

	if !reflect.DeepEqual(status, instance.Status) {
		log.Info("About to update instance status")
		instance.Status = status
		if instance.Status.Conditions == nil {
			instance.Status.Conditions = []v1alpha1.DataIndexCondition{}
		}
		if instance.Status.DependenciesStatus == nil {
			instance.Status.DependenciesStatus = []v1alpha1.DataIndexDependenciesStatus{}
		}
		if err = kubernetes.ResourceC(client).Update(instance); err != nil {
			return err
		}
	}

	return nil
}

func checkCurrentCondition(statefulSet *appsv1.StatefulSet, service *corev1.Service, resourcesUpdate bool, reconcileError bool) v1alpha1.DataIndexCondition {
	if reconcileError {
		return v1alpha1.DataIndexCondition{
			Condition:          v1alpha1.ConditionFailed,
			Message:            "Deployment Error",
			LastTransitionTime: metav1.NewTime(time.Now()),
		}
	}

	if statefulSet == nil || service == nil {
		return v1alpha1.DataIndexCondition{
			Condition:          v1alpha1.ConditionFailed,
			Message:            "Deployment Failed",
			LastTransitionTime: metav1.NewTime(time.Now()),
		}
	}

	if resourcesUpdate {
		return v1alpha1.DataIndexCondition{
			Condition:          v1alpha1.ConditionProvisioning,
			Message:            "Deploying Objects",
			LastTransitionTime: metav1.NewTime(time.Now()),
		}
	}

	if statefulSet.Status.ReadyReplicas < statefulSet.Status.Replicas {
		return v1alpha1.DataIndexCondition{
			Condition:          v1alpha1.ConditionProvisioning,
			Message:            "Deployment In Progress",
			LastTransitionTime: metav1.NewTime(time.Now()),
		}
	}

	return v1alpha1.DataIndexCondition{
		Condition:          v1alpha1.ConditionOK,
		Message:            "Deployment Finished",
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
}

func checkDependenciesStatus(instance *v1alpha1.KogitoDataIndex, client *client.Client) ([]v1alpha1.DataIndexDependenciesStatus, error) {
	// TODO: perform a real check for CRD/CRs once we have operators platform check and integration with OLM
	var deps []v1alpha1.DataIndexDependenciesStatus
	if &instance.Spec.InfinispanProperties == nil || len(instance.Spec.InfinispanProperties.URI) == 0 {
		deps = append(deps, v1alpha1.DataIndexDependenciesStatusMissingInfinispan)
	}
	if kafka, err := resource.IsKafkaServerURIResolved(instance, client); !kafka || err != nil {
		deps = append(deps, v1alpha1.DataIndexDependenciesStatusMissingKafka)
	}

	if len(deps) == 0 {
		deps = append(deps, v1alpha1.DataIndexDependenciesStatusOK)
	}

	return deps, nil
}

func getLastCondition(instance *v1alpha1.KogitoDataIndex) *v1alpha1.DataIndexCondition {
	log.Debugf("Trying to get the last condition state. Conditions are: %s", instance.Status.Conditions)
	if len(instance.Status.Conditions) > 0 {
		sort.Slice(instance.Status.Conditions, func(i, j int) bool {
			return instance.Status.Conditions[i].LastTransitionTime.Before(&instance.Status.Conditions[j].LastTransitionTime)
		})
		log.Debugf("Conditions sorted to: %s", instance.Status.Conditions)
		return &instance.Status.Conditions[len(instance.Status.Conditions)-1]
	}
	return nil
}
