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
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	imgv1 "github.com/openshift/api/image/v1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func Test_serviceDeployer_createRequiredResources_OnOCPImageStreamCreated(t *testing.T) {
	replicas := int32(1)
	instance := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.DefaultJobsServiceName,
			Namespace: t.Name(),
		},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType:       v1beta1.JobsService,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	is, tag := test.GetImageStreams(infrastructure.DefaultJobsServiceImageName, instance.Namespace, instance.Name, infrastructure.GetKogitoImageVersion())
	cli := test.NewFakeClientBuilder().OnOpenShift().AddK8sObjects(is).AddImageObjects(tag).Build()
	deployer := serviceDeployer{
		client:   cli,
		scheme:   meta.GetRegisteredSchema(),
		instance: instance,
		definition: ServiceDefinition{
			DefaultImageName: infrastructure.DefaultJobsServiceImageName,
			Request: reconcile.Request{
				NamespacedName: types.NamespacedName{Name: infrastructure.DefaultJobsServiceName, Namespace: t.Name()},
			},
		},
	}
	resources, err := deployer.createRequiredResources()
	assert.NoError(t, err)
	assert.NotEmpty(t, resources)
	// we have the Image Stream, so other resources should have been created
	assert.True(t, len(resources) > 1)
}

func Test_serviceDeployer_createRequiredResources_OnOCPNoImageStreamCreated(t *testing.T) {
	replicas := int32(1)
	instance := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.DefaultJobsServiceName,
			Namespace: t.Name(),
		},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType:       v1beta1.JobsService,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	cli := test.NewFakeClientBuilder().OnOpenShift().Build()
	deployer := serviceDeployer{
		client:   cli,
		scheme:   meta.GetRegisteredSchema(),
		instance: instance,
		definition: ServiceDefinition{
			DefaultImageName: infrastructure.DefaultJobsServiceImageName,
			Request: reconcile.Request{
				NamespacedName: types.NamespacedName{Name: infrastructure.DefaultJobsServiceName, Namespace: t.Name()},
			},
		},
	}
	resources, err := deployer.createRequiredResources()
	assert.NoError(t, err)
	assert.NotEmpty(t, resources)
	// we have the Image Stream, so other resources should have been created
	assert.True(t, len(resources) == 1)
	assert.Equal(t, resources[reflect.TypeOf(imgv1.ImageStream{})][0].GetName(), infrastructure.DefaultJobsServiceImageName)
}

func Test_serviceDeployer_createRequiredResources_CreateNewAppPropConfigMap(t *testing.T) {
	replicas := int32(1)
	kogitoKafka := test.CreateFakeKogitoKafka(t.Name())
	kogitoInfinispan := test.CreateFakeKogitoInfinispan(t.Name())

	instance := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.DefaultDataIndexName,
			Namespace: t.Name(),
		},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: v1beta1.JobsService,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Replicas: &replicas,
				Infra: []string{
					kogitoKafka.Name,
					kogitoInfinispan.Name,
				},
			},
		},
	}
	is, tag := test.GetImageStreams(infrastructure.DefaultDataIndexImageName, instance.Namespace, instance.Name, infrastructure.GetKogitoImageVersion())
	cli := test.NewFakeClientBuilder().OnOpenShift().
		AddK8sObjects(is, kogitoKafka, kogitoInfinispan).
		AddImageObjects(tag).
		Build()
	deployer := serviceDeployer{
		client:   cli,
		scheme:   meta.GetRegisteredSchema(),
		instance: instance,
		definition: ServiceDefinition{
			DefaultImageName: infrastructure.DefaultDataIndexImageName,
			Request: reconcile.Request{
				NamespacedName: types.NamespacedName{Name: infrastructure.DefaultDataIndexName, Namespace: t.Name()},
			},
		},
	}
	resources, err := deployer.createRequiredResources()
	assert.NoError(t, err)
	assert.NotEmpty(t, resources)

	assert.Equal(t, 1, len(resources[reflect.TypeOf(corev1.ConfigMap{})]))
	assert.Equal(t, infrastructure.DefaultDataIndexName+appPropConfigMapSuffix, resources[reflect.TypeOf(corev1.ConfigMap{})][0].GetName())

	assert.Equal(t, 1, len(resources[reflect.TypeOf(appsv1.Deployment{})]))
	deployment, ok := resources[reflect.TypeOf(appsv1.Deployment{})][0].(*appsv1.Deployment)
	assert.True(t, ok)
	_, ok = deployment.Spec.Template.Annotations[AppPropContentHashKey]
	assert.True(t, ok)
	_, ok = resources[reflect.TypeOf(corev1.ConfigMap{})][0].(*corev1.ConfigMap).Data[ConfigMapApplicationPropertyKey]
	assert.True(t, ok)

	// we should have TLS volumes created
	assert.Len(t, deployment.Spec.Template.Spec.Volumes, 2)                    // 1 for properties, 2 for tls
	assert.Len(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts, 2) // 1 for properties, 2 for tls
	assert.Contains(t, deployment.Spec.Template.Spec.Volumes, kogitoInfinispan.Status.Volume[0].NamedVolume.ToKubeVolume())
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts, kogitoInfinispan.Status.Volume[0].Mount)
}

func Test_serviceDeployer_createRequiredResources_CreateWithAppPropConfigMap(t *testing.T) {
	replicas := int32(1)
	instance := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.DefaultDataIndexName,
			Namespace: t.Name(),
		},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: v1beta1.JobsService,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Replicas: &replicas,
			},
		},
	}
	is, tag := test.GetImageStreams(infrastructure.DefaultDataIndexImageName, instance.Namespace, instance.Name, infrastructure.GetKogitoImageVersion())
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.DefaultDataIndexName + appPropConfigMapSuffix,
			Namespace: instance.Namespace,
		},
		Data: map[string]string{
			ConfigMapApplicationPropertyKey: defaultAppPropContent,
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(is, cm).AddImageObjects(tag).Build()
	deployer := serviceDeployer{
		client:   cli,
		scheme:   meta.GetRegisteredSchema(),
		instance: instance,
		definition: ServiceDefinition{
			DefaultImageName: infrastructure.DefaultDataIndexImageName,
			Request: reconcile.Request{
				NamespacedName: types.NamespacedName{Name: infrastructure.DefaultDataIndexName, Namespace: t.Name()},
			},
		},
	}
	resources, err := deployer.createRequiredResources()
	assert.NoError(t, err)
	assert.NotEmpty(t, resources)

	configmaps, exist := resources[reflect.TypeOf(corev1.ConfigMap{})]
	assert.True(t, exist)
	assert.Equal(t, 1, len(configmaps))
	configmaps[0].SetOwnerReferences(nil)
	assert.Equal(t, cm, configmaps[0])

	assert.Equal(t, 1, len(resources[reflect.TypeOf(appsv1.Deployment{})]))
	deployment, ok := resources[reflect.TypeOf(appsv1.Deployment{})][0].(*appsv1.Deployment)
	assert.True(t, ok)
	_, ok = deployment.Spec.Template.Annotations[AppPropContentHashKey]
	assert.True(t, ok)
}
