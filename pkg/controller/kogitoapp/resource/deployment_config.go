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

package resource

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/resource"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	dockerv10 "github.com/openshift/api/image/docker10"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultReplicas = int32(1)
	// ServiceAccountName is the name of service account used by Kogito Services Runtimes
	ServiceAccountName = "kogito-service-viewer"
)

// NewDeploymentConfig creates a new DeploymentConfig resource for the KogitoApp based on the BuildConfig runner image
func NewDeploymentConfig(kogitoApp *v1alpha1.KogitoApp, runnerBC *buildv1.BuildConfig, dockerImage *dockerv10.DockerImage) (dc *appsv1.DeploymentConfig, err error) {
	if runnerBC == nil {
		return nil, fmt.Errorf("Impossible to create the DeploymentConfig without a reference to a the service BuildConfig")
	}

	dc = &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kogitoApp.Name,
			Namespace: kogitoApp.Namespace,
		},
		Spec: appsv1.DeploymentConfigSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRolling,
			},
			Template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: kogitoApp.Name,
							// this conversion will be removed in future versions
							Env: shared.FromEnvToEnvVar(kogitoApp.Spec.Env),
							// this conversion will be removed in future versions
							Resources:       shared.FromResourcesToResourcesRequirements(kogitoApp.Spec.Resources),
							Image:           runnerBC.Spec.Output.To.Name,
							ImagePullPolicy: corev1.PullAlways,
						},
					},
					ServiceAccountName: ServiceAccountName,
				},
			},
			Triggers: appsv1.DeploymentTriggerPolicies{
				{Type: appsv1.DeploymentTriggerOnConfigChange},
				{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic:      true,
						ContainerNames: []string{kogitoApp.Name},
						From:           *runnerBC.Spec.Output.To,
					},
				},
			},
		},
	}

	meta.SetGroupVersionKind(&dc.TypeMeta, meta.KindDeploymentConfig)
	addDefaultMeta(&dc.ObjectMeta, kogitoApp)
	addDefaultMeta(&dc.Spec.Template.ObjectMeta, kogitoApp)
	addDefaultLabels(&dc.Spec.Selector, kogitoApp)
	resource.MergeImageMetadataWithDeploymentConfig(dc, dockerImage)
	resource.DiscoverPortsAndProbesFromImage(dc, dockerImage)
	setReplicas(kogitoApp, dc)

	return dc, nil
}

// setReplicas defines the number of container replicas that this DeploymentConfig will have
func setReplicas(kogitoApp *v1alpha1.KogitoApp, dc *appsv1.DeploymentConfig) {
	replicas := defaultReplicas
	if kogitoApp.Spec.Replicas != nil {
		replicas = *kogitoApp.Spec.Replicas
	}
	dc.Spec.Replicas = replicas
}
