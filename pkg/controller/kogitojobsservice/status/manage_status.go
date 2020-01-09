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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitojobsservice/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
)

var log = logger.GetLogger("jobsservice_status")

// ManageStatus handle status update for the Jobs Service
func ManageStatus(instance *v1alpha1.KogitoJobsService, cli *client.Client, errCondition error) (err error) {
	if errCondition != nil {
		instance.Status.SetFailed(v1alpha1.UnknownReason, errCondition)
		if err := update(instance, cli); err != nil {
			log.Errorf("Error while trying to set condition to error: %s", err)
			return err
		}
		// don't need to updateStatus anything else unless we break the error state
		return nil
	}

	updateStatus := false
	changed := false
	updateStatus = updateImageStatus(instance) || updateStatus
	if changed, err = updateDeploymentStatus(instance, cli); err != nil {
		return err
	}
	updateStatus = changed || updateStatus

	if changed, err = updateRouteStatus(instance, cli); err != nil {
		return err
	}
	updateStatus = changed || updateStatus

	if instance.Status.DeploymentStatus.ReadyReplicas == instance.Spec.Replicas {
		updateStatus = instance.Status.SetDeployed() || updateStatus
	} else {
		updateStatus = instance.Status.SetProvisioning() || updateStatus
	}

	if updateStatus {
		if err := update(instance, cli); err != nil {
			log.Errorf("Error while trying to update status: %s", err)
			return err
		}
	}

	return nil
}

func update(instance *v1alpha1.KogitoJobsService, cli *client.Client) error {
	err := kubernetes.ResourceC(cli).Update(instance)
	if err != nil {
		return err
	}
	return nil
}

func updateImageStatus(instance *v1alpha1.KogitoJobsService) bool {
	image := resource.NewImageResolver(instance).ResolveImage()
	if image != instance.Status.Image {
		instance.Status.Image = image
		return true
	}
	return false
}

func updateDeploymentStatus(instance *v1alpha1.KogitoJobsService, cli *client.Client) (bool, error) {
	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: instance.Name, Namespace: instance.Namespace}}
	if _, err := kubernetes.ResourceC(cli).Fetch(deployment); err != nil {
		return false, err
	}
	if !reflect.DeepEqual(instance.Status.DeploymentStatus, deployment.Status) {
		instance.Status.DeploymentStatus = deployment.Status
		return true, nil
	}
	return false, nil
}

func updateRouteStatus(instance *v1alpha1.KogitoJobsService, cli *client.Client) (bool, error) {
	if cli.IsOpenshift() {
		if exists, route, err := openshift.RouteC(cli).GetHostFromRoute(types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}); err != nil {
			return false, err
		} else if exists {
			uri := fmt.Sprintf("http://%s", route)
			if uri != instance.Status.ExternalURI {
				instance.Status.ExternalURI = uri
				return true, nil
			}
		}
	}
	return false, nil
}
