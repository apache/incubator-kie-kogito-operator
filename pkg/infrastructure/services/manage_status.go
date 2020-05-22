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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
)

// manageStatus handle status update for the Kogito Service
func (s *serviceDeployer) manageStatus(imageName string, imageTag string, errCondition error) (err error) {
	if errCondition != nil {
		s.instance.GetStatus().SetFailed(v1alpha1.UnknownReason, errCondition)
		if err := s.update(); err != nil {
			log.Errorf("Error while trying to set condition to error: %s", err)
			return err
		}
		// don't need to updateStatus anything else unless we break the error state
		return nil
	}
	var readyReplicas int32
	changed := false
	updateStatus := updateImageStatus(s.instance, imageName, imageTag, s.client)
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
		if err := s.update(); err != nil {
			log.Errorf("Error while trying to update status: %s", err)
			return err
		}
	}

	return nil
}

func updateImageStatus(instance v1alpha1.KogitoService, imageName string, imageTag string, cli *client.Client) bool {
	imageHandler := newImageHandler(instance, imageName, imageTag, cli)
	image := imageHandler.resolveRegistryImage()
	if len(image) > 0 && image != instance.GetStatus().GetImage() {
		if imageHandler.hasImageStream() {
			image = fmt.Sprintf("%s (Internal Registry)", image)
		}
		instance.GetStatus().SetImage(image)
		return true
	}
	return false
}

func updateDeploymentStatus(instance v1alpha1.KogitoService, cli *client.Client) (update bool, readyReplicas int32, err error) {
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

func updateRouteStatus(instance v1alpha1.KogitoService, cli *client.Client) (bool, error) {
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
