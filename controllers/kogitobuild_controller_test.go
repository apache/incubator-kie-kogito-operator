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

package controllers

import (
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/controllers/build"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"testing"
	"time"
)

func TestReconcileKogitoBuildSimple(t *testing.T) {
	instanceName := "quarkus-example"
	instance := &v1beta1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: t.Name()},
		Spec: v1beta1.KogitoBuildSpec{
			Type: v1beta1.RemoteSourceBuildType,
			GitSource: v1beta1.GitSource{
				URI:        "https://github.com/kiegroup/kogito-examples/",
				ContextDir: instanceName,
			},
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("250m"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("250m"),
				},
			},
		},
	}
	cli := test.NewFakeClientBuilder().OnOpenShift().AddK8sObjects(instance).Build()
	r := KogitoBuildReconciler{Client: cli, Scheme: meta.GetRegisteredSchema(), Log: logger.GetLogger("kogitoBuild reconciler")}

	// first reconciliation
	result := test.AssertReconcileMustRequeue(t, &r, instance)
	assert.Equal(t, time.Second*10, result.RequeueAfter)

	// verifying if all images have been created
	kogitoISList := &imagev1.ImageStreamList{}
	err := kubernetes.ResourceC(cli).ListWithNamespace(t.Name(), kogitoISList)
	assert.NoError(t, err)
	assert.Len(t, kogitoISList.Items, 2)
	sort.SliceStable(kogitoISList.Items, func(i, j int) bool {
		return kogitoISList.Items[i].Name < kogitoISList.Items[j].Name
	})
	assert.Equal(t, infrastructure.KogitoQuarkusJVMUbi8Image, kogitoISList.Items[0].Name)
	assert.Equal(t, infrastructure.KogitoQuarkusUbi8s2iImage, kogitoISList.Items[1].Name)
	assert.Equal(t, infrastructure.GetKogitoImageVersion(), kogitoISList.Items[0].Spec.Tags[0].Name)
	assert.Equal(t, infrastructure.GetKogitoImageVersion(), kogitoISList.Items[1].Spec.Tags[0].Name)

	// reconcile again, check for builds
	result = test.AssertReconcileMustNotRequeue(t, &r, instance)

	kogitoISList = &imagev1.ImageStreamList{}
	err = kubernetes.ResourceC(cli).ListWithNamespace(t.Name(), kogitoISList)
	assert.NoError(t, err)
	assert.Len(t, kogitoISList.Items, 4) // 2 for each the images used to buildRequest and two for the outputs
	sort.SliceStable(kogitoISList.Items, func(i, j int) bool {
		return kogitoISList.Items[i].Name < kogitoISList.Items[j].Name
	})
	assert.Equal(t, instanceName, kogitoISList.Items[2].Name)
	assert.Equal(t, build.GetBuildBuilderName(instance), kogitoISList.Items[3].Name)

	builderBC := &buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{Name: build.GetBuildBuilderName(instance), Namespace: t.Name()},
	}
	test.AssertFetchMustExist(t, cli, builderBC)

	runtimeBC := &buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: t.Name()},
	}
	test.AssertFetchMustExist(t, cli, runtimeBC)

	// reconcile one more time having everything in place
	test.AssertReconcileMustNotRequeue(t, &r, instance)

	// change something
	test.AssertFetchMustExist(t, cli, instance)
	instance.Spec.GitSource.Reference = "v1.0.0"
	err = kubernetes.ResourceC(cli).Update(instance)
	assert.NoError(t, err)

	// reconcile
	result = test.AssertReconcileMustNotRequeue(t, &r, instance)
	test.AssertFetchMustExist(t, cli, instance)

	assert.Len(t, instance.Status.Conditions, 1)
	assert.Equal(t, v1beta1.KogitoBuildRunning, instance.Status.Conditions[0].Type)
}

func TestReconcileKogitoBuildMultiple(t *testing.T) {
	kogitoServiceName := "quarkus-example"
	instanceLocalName := "quarkus-example-local"
	instanceRemote := &v1beta1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{Name: kogitoServiceName, Namespace: t.Name(), UID: test.GenerateUID()},
		Spec: v1beta1.KogitoBuildSpec{
			Type: v1beta1.RemoteSourceBuildType,
			GitSource: v1beta1.GitSource{
				URI:        "https://github.com/kiegroup/kogito-examples/",
				ContextDir: kogitoServiceName,
			},
			Runtime: v1beta1.QuarkusRuntimeType,
		},
	}
	instanceLocal := &v1beta1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{Name: instanceLocalName, Namespace: t.Name(), UID: test.GenerateUID()},
		Spec: v1beta1.KogitoBuildSpec{
			Type:                v1beta1.LocalSourceBuildType,
			Runtime:             v1beta1.QuarkusRuntimeType,
			TargetKogitoRuntime: kogitoServiceName,
		},
	}
	cli := test.NewFakeClientBuilder().OnOpenShift().AddK8sObjects(instanceRemote, instanceLocal).Build()
	r := KogitoBuildReconciler{Client: cli, Scheme: meta.GetRegisteredSchema(), Log: logger.GetLogger("kogitoBuild reconciler")}

	// first reconciliation
	result := test.AssertReconcileMustRequeue(t, &r, instanceRemote)
	assert.Equal(t, time.Second*10, result.RequeueAfter)
	// we won't requeue since the Kogito ImageStreams should be created for the first instance
	result = test.AssertReconcileMustNotRequeue(t, &r, instanceLocal)
	// now we create the objects for Remote
	result = test.AssertReconcileMustNotRequeue(t, &r, instanceRemote)

	is, err := services.GetSharedDeployedImageStream(kogitoServiceName, t.Name(), cli)
	assert.NoError(t, err)
	assert.NotNil(t, is)
	assert.Len(t, is.OwnerReferences, 2) // we have two owners!
	// and none of them are the controller, this is an anarchy!
	for _, owner := range is.OwnerReferences {
		assert.False(t, *owner.Controller)
	}
}
