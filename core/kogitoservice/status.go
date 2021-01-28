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

package kogitoservice

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
)

// StatusHandler ...
type StatusHandler interface {
	HandleStatusUpdate(instance api.KogitoService, err *error)
}

type statusHandler struct {
	client *client.Client
	log    logger.Logger
}

// NewStatusHandler ...
func NewStatusHandler(client *client.Client, log logger.Logger) StatusHandler {
	return &statusHandler{
		client: client,
		log:    log,
	}
}

func (s *statusHandler) HandleStatusUpdate(instance api.KogitoService, err *error) {
	s.log.Info("Updating status for Kogito Service")
	if statusErr := s.ensureResourcesStatusChanges(instance, *err); statusErr != nil {
		s.log.Error(statusErr, "Error while updating Status for Kogito Service")
		return
	}
	s.log.Info("Finished Kogito Service reconciliation")
}

func (s *statusHandler) ensureResourcesStatusChanges(instance api.KogitoService, errCondition error) (err error) {
	if errCondition != nil {
		instance.GetStatus().SetFailed(reasonForError(errCondition), errCondition)
		if err := s.updateStatus(instance); err != nil {
			s.log.Error(err, "Error while trying to set condition to error")
			return err
		}
		// don't need to update anything else or we break the error state
		return nil
	}
	var readyReplicas int32
	changed := false
	updateStatus, err := s.updateImageStatus(instance)
	if err != nil {
		return err
	}
	if changed, readyReplicas, err = s.updateDeploymentStatus(instance); err != nil {
		return err
	}
	updateStatus = changed || updateStatus

	if changed, err = s.updateRouteStatus(instance); err != nil {
		return err
	}
	updateStatus = changed || updateStatus

	if readyReplicas == *instance.GetSpec().GetReplicas() && readyReplicas > 0 {
		updateStatus = instance.GetStatus().SetDeployed() || updateStatus
	} else {
		updateStatus = instance.GetStatus().SetProvisioning() || updateStatus
	}

	if updateStatus {
		if err := s.updateStatus(instance); err != nil {
			s.log.Error(err, "Error while trying to update status")
			return err
		}
	}
	return nil
}

func (s *statusHandler) updateStatus(instance api.KogitoService) error {
	// Sanity check since the Status CR needs a reference for the object
	if instance.GetStatus() != nil && instance.GetStatus().GetConditions() == nil {
		instance.GetStatus().SetConditions([]api.Condition{})
	}
	err := kubernetes.ResourceC(s.client).UpdateStatus(instance)
	if err != nil {
		return err
	}
	return nil
}

func (s *statusHandler) updateImageStatus(instance api.KogitoService) (bool, error) {
	deploymentHandler := infrastructure.NewDeploymentHandler(s.client, s.log)
	deployment, err := deploymentHandler.MustFetchDeployment(types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})
	if err != nil {
		return false, err
	}
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		image := deployment.Spec.Template.Spec.Containers[0].Image
		if len(image) > 0 && image != instance.GetStatus().GetImage() {
			instance.GetStatus().SetImage(deployment.Spec.Template.Spec.Containers[0].Image)
			return true, nil
		}
	}
	return false, nil
}

func (s *statusHandler) updateDeploymentStatus(instance api.KogitoService) (update bool, readyReplicas int32, err error) {
	deploymentHandler := infrastructure.NewDeploymentHandler(s.client, s.log)
	deployment, err := deploymentHandler.MustFetchDeployment(types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})
	if err != nil {
		return false, 0, err
	}

	if !reflect.DeepEqual(instance.GetStatus().GetDeploymentConditions(), deployment.Status.Conditions) {
		instance.GetStatus().SetDeploymentConditions(deployment.Status.Conditions)
		return true, deployment.Status.ReadyReplicas, nil
	}
	return false, deployment.Status.ReadyReplicas, nil
}

func (s *statusHandler) updateRouteStatus(instance api.KogitoService) (bool, error) {
	if s.client.IsOpenshift() {
		routeHandler := infrastructure.NewRouteHandler(s.client, s.log)
		route, err := routeHandler.GetHostFromRoute(types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})
		if err != nil {
			return false, err
		}

		if len(route) > 0 {
			uri := fmt.Sprintf("http://%s", route)
			if uri != instance.GetStatus().GetExternalURI() {
				instance.GetStatus().SetExternalURI(uri)
				return true, nil
			}
		}
	}
	return false, nil
}
