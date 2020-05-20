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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	imgv1 "github.com/openshift/api/image/v1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
	"time"
)

func Test_serviceDeployer_createRequiredResources_OnOCPImageStreamCreated(t *testing.T) {
	replicas := int32(1)
	instance := &v1alpha1.KogitoJobsService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.DefaultJobsServiceName,
			Namespace: t.Name(),
		},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
			InfinispanMeta: v1alpha1.InfinispanMeta{
				InfinispanProperties: v1alpha1.InfinispanConnectionProperties{
					UseKogitoInfra: false,
					URI:            "another-uri:11222",
				},
			},
		},
	}
	is, tag := test.GetImageStreams(infrastructure.DefaultJobsServiceImageName, instance.Namespace, instance.Name, infrastructure.GetRuntimeImageVersion())
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance, is}, []runtime.Object{tag}, nil)
	deployer := serviceDeployer{
		client:       cli,
		scheme:       meta.GetRegisteredSchema(),
		instanceList: &v1alpha1.KogitoJobsServiceList{},
		instance:     instance,
		definition: ServiceDefinition{
			DefaultImageName: infrastructure.DefaultJobsServiceImageName,
			Request: reconcile.Request{
				NamespacedName: types.NamespacedName{Name: infrastructure.DefaultJobsServiceName, Namespace: t.Name()},
			},
		},
	}
	resources, reconcileAfter, err := deployer.createRequiredResources()
	assert.NoError(t, err)
	assert.NotEmpty(t, resources)
	// we have the Image Stream, so other resources should have been created
	assert.True(t, len(resources) > 1)
	assert.Equal(t, reconcileAfter, time.Duration(0))
}

func Test_serviceDeployer_createRequiredResources_OnOCPNoImageStreamCreated(t *testing.T) {
	replicas := int32(1)
	instance := &v1alpha1.KogitoJobsService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.DefaultJobsServiceName,
			Namespace: t.Name(),
		},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
			InfinispanMeta: v1alpha1.InfinispanMeta{
				InfinispanProperties: v1alpha1.InfinispanConnectionProperties{
					UseKogitoInfra: false,
					URI:            "another-uri:11222",
				},
			},
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance}, nil, nil)
	deployer := serviceDeployer{
		client:       cli,
		scheme:       meta.GetRegisteredSchema(),
		instanceList: &v1alpha1.KogitoJobsServiceList{},
		instance:     instance,
		definition: ServiceDefinition{
			DefaultImageName: infrastructure.DefaultJobsServiceImageName,
			Request: reconcile.Request{
				NamespacedName: types.NamespacedName{Name: infrastructure.DefaultJobsServiceName, Namespace: t.Name()},
			},
		},
	}
	resources, reconcileAfter, err := deployer.createRequiredResources()
	assert.NoError(t, err)
	assert.NotEmpty(t, resources)
	assert.Equal(t, reconcileAfter, time.Duration(0))
	// we have the Image Stream, so other resources should have been created
	assert.True(t, len(resources) == 1)
	assert.Equal(t, resources[reflect.TypeOf(imgv1.ImageStream{})][0].GetName(), infrastructure.DefaultJobsServiceImageName)
}

func Test_serviceDeployer_createRequiredResources_RequiresDataIndex(t *testing.T) {
	replicas := int32(1)
	instance := &v1alpha1.KogitoMgmtConsole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.DefaultMgmtConsoleName,
			Namespace: t.Name(),
		},
		Spec: v1alpha1.KogitoMgmtConsoleSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	is, tag := test.GetImageStreams(infrastructure.DefaultMgmtConsoleImageName, instance.Namespace, instance.Name, infrastructure.GetRuntimeImageVersion())
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance, is}, []runtime.Object{tag}, nil)
	deployer := serviceDeployer{
		client:       cli,
		scheme:       meta.GetRegisteredSchema(),
		instanceList: &v1alpha1.KogitoJobsServiceList{},
		instance:     instance,
		definition: ServiceDefinition{
			DefaultImageName:  infrastructure.DefaultMgmtConsoleImageName,
			RequiresDataIndex: true,
			Request: reconcile.Request{
				NamespacedName: types.NamespacedName{Name: infrastructure.DefaultMgmtConsoleName, Namespace: t.Name()},
			},
		},
	}
	resources, reconcileAfter, err := deployer.createRequiredResources()
	assert.NoError(t, err)
	assert.NotEmpty(t, resources)
	// we have the Image Stream, so other resources should have been created
	assert.True(t, len(resources) > 1)
	// we don't have data index set
	assert.Equal(t, reconcileAfter, dataIndexDependencyReconcileAfter)
}

func Test_serviceDeployer_createRequiredResources_CreateNewAppPropConfigMap(t *testing.T) {
	replicas := int32(1)
	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.DefaultDataIndexName,
			Namespace: t.Name(),
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	is, tag := test.GetImageStreams(infrastructure.DefaultDataIndexImageName, instance.Namespace, instance.Name, infrastructure.GetRuntimeImageVersion())
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance, is}, []runtime.Object{tag}, nil)
	deployer := serviceDeployer{
		client:       cli,
		scheme:       meta.GetRegisteredSchema(),
		instanceList: &v1alpha1.KogitoDataIndexList{},
		instance:     instance,
		definition: ServiceDefinition{
			DefaultImageName: infrastructure.DefaultDataIndexImageName,
			Request: reconcile.Request{
				NamespacedName: types.NamespacedName{Name: infrastructure.DefaultDataIndexName, Namespace: t.Name()},
			},
		},
	}
	resources, _, err := deployer.createRequiredResources()
	assert.NoError(t, err)
	assert.NotEmpty(t, resources)

	assert.Equal(t, 1, len(resources[reflect.TypeOf(corev1.ConfigMap{})]))
	assert.Equal(t, infrastructure.DefaultDataIndexName+appPropConfigMapSuffix, resources[reflect.TypeOf(corev1.ConfigMap{})][0].GetName())

	assert.Equal(t, 1, len(resources[reflect.TypeOf(appsv1.Deployment{})]))
	deployment, ok := resources[reflect.TypeOf(appsv1.Deployment{})][0].(*appsv1.Deployment)
	assert.True(t, ok)
	_, ok = deployment.Spec.Template.Annotations[AppPropContentHashKey]
	assert.True(t, ok)
}

func Test_serviceDeployer_createRequiredResources_CreateNoAppPropConfigMap(t *testing.T) {
	replicas := int32(1)
	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.DefaultDataIndexName,
			Namespace: t.Name(),
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	is, tag := test.GetImageStreams(infrastructure.DefaultDataIndexImageName, instance.Namespace, instance.Name, infrastructure.GetRuntimeImageVersion())
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.DefaultDataIndexName + appPropConfigMapSuffix,
			Namespace: instance.Namespace,
		},
		Data: map[string]string{
			appPropFileName: defaultAppPropContent,
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance, is, cm}, []runtime.Object{tag}, nil)
	deployer := serviceDeployer{
		client:       cli,
		scheme:       meta.GetRegisteredSchema(),
		instanceList: &v1alpha1.KogitoDataIndexList{},
		instance:     instance,
		definition: ServiceDefinition{
			DefaultImageName: infrastructure.DefaultDataIndexImageName,
			Request: reconcile.Request{
				NamespacedName: types.NamespacedName{Name: infrastructure.DefaultDataIndexName, Namespace: t.Name()},
			},
		},
	}
	resources, _, err := deployer.createRequiredResources()
	assert.NoError(t, err)
	assert.NotEmpty(t, resources)

	_, exist := resources[reflect.TypeOf(corev1.ConfigMap{})]
	assert.False(t, exist)

	assert.Equal(t, 1, len(resources[reflect.TypeOf(appsv1.Deployment{})]))
	deployment, ok := resources[reflect.TypeOf(appsv1.Deployment{})][0].(*appsv1.Deployment)
	assert.True(t, ok)
	_, ok = deployment.Spec.Template.Annotations[AppPropContentHashKey]
	assert.True(t, ok)
}
