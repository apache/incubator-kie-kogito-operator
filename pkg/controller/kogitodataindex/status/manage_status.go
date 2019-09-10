package status

import (
	"reflect"
	"sort"
	"time"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitodataindex/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var log = logger.GetLogger("status_kogitodataindex")

// ManageStatus will garantee the status changes
func ManageStatus(instance *v1alpha1.KogitoDataIndex, resources *resource.KogitoDataIndexResources, client *client.Client) error {
	var err error
	status := v1alpha1.KogitoDataIndexStatus{}
	currentCondition := v1alpha1.DataIndexCondition{}

	if resources.StatefulSet != nil {
		status.DeploymentStatus = resources.StatefulSet.Status
	}

	if resources.Service != nil {
		status.ServiceStatus = resources.Service.Status
	}

	if status.DependenciesStatus, err = checkDependenciesStatus(instance); err != nil {
		return err
	}

	if resources.Route != nil {
		log.Debugf("Trying to get the host for the route %s", resources.Route.Name)
		if status.Route, err =
			openshift.RouteC(client).GetHostFromRoute(
				types.NamespacedName{Name: resources.Route.Name, Namespace: resources.Route.Namespace}); err != nil {
			return err
		}
	} else {
		log.Debugf("Route is nil, impossible to get host to set in the status", resources.Route)
	}

	status.Conditions = instance.Status.Conditions
	if currentCondition, err = checkCurrentCondition(resources, client); err != nil {
		return err
	}

	lastCondition := getLastCondition(instance)
	if lastCondition == nil || (currentCondition.Condition != lastCondition.Condition && currentCondition.Message != lastCondition.Message) {
		log.Debugf("Creating new status conditions. Actual conditions: %s. Current condition: %s", instance.Status.Conditions, currentCondition)
		if &status.Conditions == nil {
			status.Conditions = []v1alpha1.DataIndexCondition{}
		}
		status.Conditions = append(status.Conditions, currentCondition)
	}

	if !reflect.DeepEqual(status, instance.Status) {
		log.Debugf("About to update intance status")
		instance.Status = status
		if err = kubernetes.ResourceC(client).UpdateStatus(instance); err != nil {
			return err
		}
	}

	return nil
}

func checkCurrentCondition(resources *resource.KogitoDataIndexResources, client *client.Client) (v1alpha1.DataIndexCondition, error) {
	if resources.StatefulSet == nil ||
		resources.ProtoBufConfigMap == nil ||
		resources.Service == nil {
		return v1alpha1.DataIndexCondition{
			Condition:          v1alpha1.ConditionProvisioning,
			Message:            "Not all objects created",
			LastTransitionTime: metav1.NewTime(time.Now()),
		}, nil
	}

	if resources.StatefulSet.Status.ReadyReplicas == resources.StatefulSet.Status.Replicas {
		return v1alpha1.DataIndexCondition{
			Condition:          v1alpha1.ConditionOK,
			Message:            "Deployment Finished",
			LastTransitionTime: metav1.NewTime(time.Now()),
		}, nil
	}

	return v1alpha1.DataIndexCondition{
		Condition:          v1alpha1.ConditionProvisioning,
		Message:            "Deployment Not Started",
		LastTransitionTime: metav1.NewTime(time.Now()),
	}, nil
}

func checkDependenciesStatus(instance *v1alpha1.KogitoDataIndex) ([]v1alpha1.DataIndexDependenciesStatus, error) {
	// TODO: perform a real check for CRD/CRs once we have operators platform check and integration with OLM
	deps := []v1alpha1.DataIndexDependenciesStatus{}
	if &instance.Spec.Infinispan == nil || len(instance.Spec.Infinispan.ServiceURI) == 0 {
		deps = append(deps, v1alpha1.DataIndexDependenciesStatusMissingInfinispan)
	}
	if &instance.Spec.Kafka == nil || len(instance.Spec.Kafka.ServiceURI) == 0 {
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
