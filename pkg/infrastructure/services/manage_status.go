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

package services

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
)

func (s *serviceDeployer) handleStatusUpdate(instance v1beta1.KogitoService, err *error) {
	log.Infof("Updating status for Kogito Service %s", instance.GetName())
	if statusErr := s.ensureResourcesStatusChanges(*err); statusErr != nil {
		log.Errorf("Error while updating Status for Kogito Service: %v", statusErr)
		return
	}
	log.Infof("Finished Kogito Service %s reconciliation", instance.GetName())
}

func (s *serviceDeployer) ensureResourcesStatusChanges(errCondition error) (err error) {
	if errCondition != nil {
		s.instance.GetStatus().SetFailed(reasonForError(errCondition), errCondition)
		if err := s.updateStatus(); err != nil {
			log.Errorf("Error while trying to set condition to error: %s", err)
			return err
		}
		// don't need to update anything else or we break the error state
		return nil
	}
	var readyReplicas int32
	changed := false
	updateStatus, err := updateImageStatus(s.instance, s.client)
	if err != nil {
		return err
	}
	if changed, readyReplicas, err = updateDeploymentStatus(s.instance, s.client); err != nil {
		return err
	}
	updateStatus = changed || updateStatus

	if changed, err = updateRouteStatus(s.instance, s.client); err != nil {
		return err
	}
	updateStatus = changed || updateStatus

	if readyReplicas == *s.instance.GetSpec().GetReplicas() && readyReplicas > 0 {
		updateStatus = s.instance.GetStatus().SetDeployed() || updateStatus
	} else {
		updateStatus = s.instance.GetStatus().SetProvisioning() || updateStatus
	}

	if updateStatus {
		if err := s.updateStatus(); err != nil {
			log.Errorf("Error while trying to update status: %s", err)
			return err
		}
	}

	return nil
}

func (s *serviceDeployer) updateStatus() error {
	// Sanity check since the Status CR needs a reference for the object
	if s.instance.GetStatus() != nil && s.instance.GetStatus().GetConditions() == nil {
		s.instance.GetStatus().SetConditions([]v1beta1.Condition{})
	}
	err := kubernetes.ResourceC(s.client).UpdateStatus(s.instance)
	if err != nil {
		return err
	}
	return nil
}

func updateImageStatus(instance v1beta1.KogitoService, cli *client.Client) (bool, error) {
	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: instance.GetName(), Namespace: instance.GetNamespace()}}
	exists, err := kubernetes.ResourceC(cli).Fetch(deployment)
	if err != nil {
		return false, err
	}
	if exists && len(deployment.Spec.Template.Spec.Containers) > 0 {
		image := deployment.Spec.Template.Spec.Containers[0].Image
		if len(image) > 0 && image != instance.GetStatus().GetImage() {
			instance.GetStatus().SetImage(deployment.Spec.Template.Spec.Containers[0].Image)
			return true, nil
		}
	}
	return false, nil
}

func updateDeploymentStatus(instance v1beta1.KogitoService, cli *client.Client) (update bool, readyReplicas int32, err error) {
	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: instance.GetName(), Namespace: instance.GetNamespace()}}
	if _, err := kubernetes.ResourceC(cli).Fetch(deployment); err != nil {
		return false, 0, err
	}
	if !reflect.DeepEqual(instance.GetStatus().GetDeploymentConditions(), deployment.Status.Conditions) {
		instance.GetStatus().SetDeploymentConditions(deployment.Status.Conditions)
		return true, deployment.Status.ReadyReplicas, nil
	}
	return false, deployment.Status.ReadyReplicas, nil
}

func updateRouteStatus(instance v1beta1.KogitoService, cli *client.Client) (bool, error) {
	if cli.IsOpenshift() {
		if exists, route, err := openshift.RouteC(cli).GetHostFromRoute(types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()}); err != nil {
			return false, err
		} else if exists {
			uri := fmt.Sprintf("http://%s", route)
			if uri != instance.GetStatus().GetExternalURI() {
				instance.GetStatus().SetExternalURI(uri)
				return true, nil
			}
		}
	}
	return false, nil
}
