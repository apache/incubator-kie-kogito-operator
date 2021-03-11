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
	"github.com/kiegroup/kogito-cloud-operator/api"
	"github.com/kiegroup/kogito-cloud-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/core/framework"
	"github.com/kiegroup/kogito-cloud-operator/core/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/core/operator"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
)

// StatusHandler ...
type StatusHandler interface {
	HandleStatusUpdate(instance api.KogitoService, err *error)
}

type statusHandler struct {
	*operator.Context
}

// NewStatusHandler ...
func NewStatusHandler(context *operator.Context) StatusHandler {
	return &statusHandler{
		context,
	}
}

func (s *statusHandler) HandleStatusUpdate(instance api.KogitoService, err *error) {
	s.Log.Info("Updating status for Kogito Service")
	if statusErr := s.ensureResourcesStatusChanges(instance, *err); statusErr != nil {
		s.Log.Error(statusErr, "Error while updating Status for Kogito Service")
		return
	}
	s.Log.Info("Finished Kogito Service reconciliation")
}

func (s *statusHandler) ensureResourcesStatusChanges(instance api.KogitoService, errCondition error) (err error) {
	if errCondition != nil {
		s.SetFailed(instance.GetStatus(), reasonForError(errCondition), errCondition)
		if err := s.updateStatus(instance); err != nil {
			s.Log.Error(err, "Error while trying to set condition to error")
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
		updateStatus = s.SetDeployed(instance.GetStatus()) || updateStatus
	} else {
		updateStatus = s.SetProvisioning(instance.GetStatus()) || updateStatus
	}

	if updateStatus {
		if err := s.updateStatus(instance); err != nil {
			s.Log.Error(err, "Error while trying to update status")
			return err
		}
	}
	return nil
}

func (s *statusHandler) updateStatus(instance api.KogitoService) error {
	// Sanity check since the Status CR needs a reference for the object
	if instance.GetStatus() != nil && instance.GetStatus().GetConditions() == nil {
		conditions := make([]api.ConditionInterface, 1)
		instance.GetStatus().SetConditions(conditions)
	}
	err := kubernetes.ResourceC(s.Client).UpdateStatus(instance)
	if err != nil {
		return err
	}
	return nil
}

func (s *statusHandler) updateImageStatus(instance api.KogitoService) (bool, error) {
	deploymentHandler := infrastructure.NewDeploymentHandler(s.Context)
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
	deploymentHandler := infrastructure.NewDeploymentHandler(s.Context)
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
	if s.Client.IsOpenshift() {
		routeHandler := infrastructure.NewRouteHandler(s.Context)
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

// SetDeployed Updates the condition with the DeployedCondition and True status
func (s *statusHandler) SetDeployed(c api.ConditionMetaInterface) bool {
	size := len(c.GetConditions())
	if size > 0 && c.GetConditions()[size-1].GetType() == api.DeployedConditionType {
		return false
	}
	condition := c.NewDeployedCondition()
	c.SetConditions(s.addCondition(c, condition))
	return true
}

// SetProvisioning Sets the condition type to Provisioning and status True if not yet set.
func (s *statusHandler) SetProvisioning(c api.ConditionMetaInterface) bool {
	size := len(c.GetConditions())
	if size > 0 && c.GetConditions()[size-1].GetType() == api.ProvisioningConditionType {
		return false
	}
	condition := c.NewProvisioningCondition()
	c.SetConditions(s.addCondition(c, condition))
	return true
}

// SetFailed Sets the failed condition with the error reason and message
func (s *statusHandler) SetFailed(c api.ConditionMetaInterface, reason api.KogitoServiceConditionReason, err error) {
	condition := c.NewFailedCondition(reason, err)
	c.SetConditions(s.addCondition(c, condition))
}

// addCondition adds a condition to the condition array ensuring the max buffer
func (s *statusHandler) addCondition(c api.ConditionMetaInterface, condition api.ConditionInterface) []api.ConditionInterface {
	conditionCollection := framework.NewConditionCollection(c.GetConditions())
	return conditionCollection.AddCondition(condition)
}
