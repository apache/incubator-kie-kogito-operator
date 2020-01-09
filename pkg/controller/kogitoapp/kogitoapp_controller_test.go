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

package kogitoapp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"reflect"
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	cachev1 "sigs.k8s.io/controller-runtime/pkg/cache/informertest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kogitoclient "github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	kogitores "github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/status"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	dockerv10 "github.com/openshift/api/image/docker10"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"

	"github.com/stretchr/testify/assert"
)

var (
	cpuResource    = v1alpha1.ResourceCPU
	memoryResource = v1alpha1.ResourceMemory
	cpuValue       = "1"
	cr             = v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Resources: v1alpha1.Resources{
				Limits: []v1alpha1.ResourceMap{
					{
						Resource: cpuResource,
						Value:    cpuValue,
					},
				},
			},
		},
	}
)

func createFakeKogitoApp() *v1alpha1.KogitoApp {
	gitURL := "https://github.com/kiegroup/kogito-examples/"
	kogitoapp := &cr
	kogitoapp.Spec.Build = &v1alpha1.KogitoAppBuildObject{
		ImageS2I: v1alpha1.ImageStream{
			ImageStreamTag: "0.4.0",
		},
		ImageRuntime: v1alpha1.ImageStream{
			ImageStreamTag: "0.4.0",
		},
		GitSource: &v1alpha1.GitSource{
			URI:        &gitURL,
			ContextDir: "jbpm-quarkus-example",
		},
	}

	return kogitoapp
}

func createFakeImages(kogitoAppName string, runtimeLabels map[string]string) []runtime.Object {
	if nil == runtimeLabels {
		runtimeLabels = map[string]string{}
		runtimeLabels[openshift.ImageLabelForExposeServices] = "8080:http"
	}
	dockerImageRaw, _ := json.Marshal(&dockerv10.DockerImage{
		Config: &dockerv10.DockerConfig{
			Labels: runtimeLabels,
		},
	})

	isTag := imgv1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s:latest", kogitoAppName),
			Namespace: "test",
		},
		Image: imgv1.Image{
			DockerImageMetadata: runtime.RawExtension{
				Raw: dockerImageRaw,
			},
		},
	}

	isTagBuild := imgv1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s:latest", kogitoAppName+"-build"),
			Namespace: "test",
		},
		Image: imgv1.Image{
			DockerImageMetadata: runtime.RawExtension{
				Raw: dockerImageRaw,
			},
		},
	}

	image1 := imgv1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s:%s", kogitores.KogitoQuarkusUbi8Image, "0.4.0"),
			Namespace: "test",
		},
		Image: imgv1.Image{
			DockerImageMetadata: runtime.RawExtension{
				Raw: dockerImageRaw,
			},
		},
	}
	image2 := imgv1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s:%s", kogitores.KogitoQuarkusJVMUbi8Image, "0.4.0"),
			Namespace: "test",
		},
		Image: imgv1.Image{
			DockerImageMetadata: runtime.RawExtension{
				Raw: dockerImageRaw,
			},
		},
	}
	image3 := imgv1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s:%s", kogitores.KogitoQuarkusUbi8s2iImage, "0.4.0"),
			Namespace: "test",
		},
		Image: imgv1.Image{
			DockerImageMetadata: runtime.RawExtension{
				Raw: dockerImageRaw,
			},
		},
	}
	image4 := imgv1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s:%s", kogitores.KogitoSpringbootUbi8Image, "0.4.0"),
			Namespace: "test",
		},
		Image: imgv1.Image{
			DockerImageMetadata: runtime.RawExtension{
				Raw: dockerImageRaw,
			},
		},
	}
	image5 := imgv1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s:%s", kogitores.KogitoSpringbootUbi8s2iImage, "0.4.0"),
			Namespace: "test",
		},
		Image: imgv1.Image{
			DockerImageMetadata: runtime.RawExtension{
				Raw: dockerImageRaw,
			},
		},
	}
	image6 := imgv1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s:%s", kogitores.KogitoDataIndexImage, "0.4.0"),
			Namespace: "test",
		},
		Image: imgv1.Image{
			DockerImageMetadata: runtime.RawExtension{
				Raw: dockerImageRaw,
			},
		},
	}

	return []runtime.Object{&isTag, &isTagBuild, &image1, &image2, &image3, &image4, &image5, &image6}
}

func TestNewContainerWithResource(t *testing.T) {
	container := corev1.Container{
		Name:            cr.Name,
		Env:             shared.FromEnvToEnvVar(cr.Spec.Env),
		Resources:       shared.FromResourcesToResourcesRequirements(cr.Spec.Resources),
		ImagePullPolicy: corev1.PullAlways,
	}
	assert.NotNil(t, container)
	cpuQty := resource.MustParse(cpuValue)
	assert.Equal(t, container.Resources.Limits.Cpu(), &cpuQty)
	assert.Equal(t, container.Resources.Requests.Cpu(), &resource.Quantity{Format: resource.DecimalSI})
}

func TestKogitoAppWithResource(t *testing.T) {
	kogitoapp := createFakeKogitoApp()
	images := createFakeImages(kogitoapp.Name, nil)
	objs := []runtime.Object{kogitoapp}
	fakeClient := test.CreateFakeClient(objs, images, []runtime.Object{})

	// ********** sanity check
	kogitoAppList := &v1alpha1.KogitoAppList{}
	err := fakeClient.ControlCli.List(context.TODO(), kogitoAppList, client.InNamespace("test"))
	if err != nil {
		t.Fatalf("Failed to list kogitoapp (%v)", err)
	}
	assert.True(t, len(kogitoAppList.Items) > 0)
	assert.True(t, kogitoAppList.Items[0].Spec.Resources.Limits[0].Resource == cpuResource)
	fakeCache := &cachev1.FakeInformers{}
	// call reconcile object and mock image and build clients
	r := &ReconcileKogitoApp{
		client: fakeClient,
		scheme: meta.GetRegisteredSchema(),
		cache:  fakeCache,
	}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      kogitoapp.Name,
			Namespace: kogitoapp.Namespace,
		},
	}

	_, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// Let's verify if the objects have been built
	dc := &appsv1.DeploymentConfig{}
	_, err = kubernetes.ResourceC(r.client).FetchWithKey(types.NamespacedName{Name: kogitoapp.Name, Namespace: kogitoapp.Namespace}, dc)
	assert.NoError(t, err)
	assert.NotNil(t, dc)
	assert.Len(t, dc.Spec.Template.Spec.Containers, 1)
	assert.Len(t, dc.GetOwnerReferences(), 1)

	bcS2I := &buildv1.BuildConfig{}
	_, err = kubernetes.ResourceC(r.client).FetchWithKey(types.NamespacedName{Name: kogitoapp.Name + kogitores.BuildS2INameSuffix, Namespace: kogitoapp.Namespace}, bcS2I)
	assert.NoError(t, err)
	assert.NotNil(t, bcS2I)

	assert.Equal(t, "0", bcS2I.Spec.Resources.Limits.Cpu().String())
	assert.Equal(t, "0", bcS2I.Spec.Resources.Limits.Memory().String())

	for _, isName := range kogitores.ImageStreamNameList {
		hasIs, _ := openshift.ImageStreamC(r.client).FetchTag(types.NamespacedName{Name: isName, Namespace: "test"}, "0.4.0")
		assert.NotNil(t, hasIs)
	}

}

func TestReconcileKogitoApp_updateKogitoAppStatus(t *testing.T) {
	kogitoapp := createFakeKogitoApp()
	kogitoapp.Status = v1alpha1.KogitoAppStatus{
		Builds: v1alpha1.Builds{
			Complete: []string{
				"test-app",
				"test-app-build",
			},
		},
		Route: "http://localhost",
		ConditionsMeta: v1alpha1.ConditionsMeta{
			Conditions: []v1alpha1.Condition{
				{
					Type: v1alpha1.DeployedConditionType,
				},
			},
		},
	}
	s := meta.GetRegisteredSchema()
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      kogitoapp.Name,
			Namespace: kogitoapp.Namespace,
		},
	}
	buildconfigRuntime := &buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kogitoapp.Name,
			Namespace: kogitoapp.Namespace,
			Labels: map[string]string{
				kogitores.LabelKeyAppName:   kogitoapp.Name,
				kogitores.LabelKeyBuildType: string(kogitores.BuildTypeRuntime),
			},
		},
	}
	buildconfigS2I := &buildv1.BuildConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kogitoapp.Name + "-build",
			Namespace: kogitoapp.Namespace,
			Labels: map[string]string{
				kogitores.LabelKeyAppName:   kogitoapp.Name,
				kogitores.LabelKeyBuildType: string(kogitores.BuildTypeS2I),
			},
		},
	}
	buildList := &buildv1.BuildList{
		Items: []buildv1.Build{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      kogitoapp.Name,
					Namespace: kogitoapp.Namespace,
					Labels: map[string]string{
						kogitores.LabelKeyAppName:   kogitoapp.Name,
						kogitores.LabelKeyBuildType: string(kogitores.BuildTypeRuntime),
					},
				},
				Status: buildv1.BuildStatus{
					Phase: buildv1.BuildPhaseComplete,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      kogitoapp.Name + "-build",
					Namespace: kogitoapp.Namespace,
					Labels: map[string]string{
						kogitores.LabelKeyAppName:   kogitoapp.Name,
						kogitores.LabelKeyBuildType: string(kogitores.BuildTypeS2I),
					},
				},
				Status: buildv1.BuildStatus{
					Phase: buildv1.BuildPhaseComplete,
				},
			},
		},
	}
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kogitoapp.Name,
			Namespace: kogitoapp.Namespace,
		},
		Spec: routev1.RouteSpec{
			Host: "localhost",
		},
	}

	runtimeObjs := []runtime.Object{kogitoapp, route}
	imageObjs := createFakeImages(kogitoapp.Name, nil)
	buildObjs := []runtime.Object{buildconfigRuntime, buildconfigS2I, buildList}

	clientfake := test.CreateFakeClient(runtimeObjs, imageObjs, buildObjs)

	kogitoAppResources := &kogitores.KogitoAppResources{
		BuildConfigRuntime: buildconfigRuntime,
		BuildConfigS2I:     buildconfigS2I,
		DeploymentConfig: &appsv1.DeploymentConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      kogitoapp.Name,
				Namespace: kogitoapp.Namespace,
			},
		},
		Route: route,
	}

	type fields struct {
		client *kogitoclient.Client
		scheme *runtime.Scheme
		cache  cache.Cache
	}
	type args struct {
		request              *reconcile.Request
		instance             *v1alpha1.KogitoApp
		kogitoResources      *kogitores.KogitoAppResources
		updateResourceResult *status.UpdateResourcesResult
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		result *reconcile.Result
		err    error
	}{
		{
			"Success",
			fields{
				clientfake,
				s,
				mockCache{
					kogitoApp: kogitoapp,
				},
			},
			args{
				&req,
				kogitoapp,
				kogitoAppResources,
				&status.UpdateResourcesResult{},
			},
			&reconcile.Result{},
			nil,
		},
		{
			"Error",
			fields{
				clientfake,
				s,
				&cachev1.FakeInformers{},
			},
			args{
				&req,
				kogitoapp,
				kogitoAppResources,
				&status.UpdateResourcesResult{
					Err: fmt.Errorf("error"),
				},
			},
			&reconcile.Result{},
			fmt.Errorf("error"),
		},
		{
			"Updated",
			fields{
				clientfake,
				s,
				&cachev1.FakeInformers{},
			},
			args{
				&req,
				kogitoapp,
				kogitoAppResources,
				&status.UpdateResourcesResult{},
			},
			&reconcile.Result{},
			nil,
		},
		{
			"RequeueAfter",
			fields{
				test.CreateFakeClient([]runtime.Object{kogitoapp}, imageObjs, buildObjs),
				s,
				mockCache{
					kogitoApp: kogitoapp,
				},
			},
			args{
				&req,
				kogitoapp,
				kogitoAppResources,
				&status.UpdateResourcesResult{},
			},
			&reconcile.Result{RequeueAfter: time.Duration(500) * time.Millisecond},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReconcileKogitoApp{
				client: tt.fields.client,
				scheme: tt.fields.scheme,
				cache:  tt.fields.cache,
			}

			result := &reconcile.Result{}
			var err error

			r.updateKogitoAppStatus(tt.args.request, tt.args.instance, tt.args.kogitoResources, tt.args.updateResourceResult, result, &err)

			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("updateKogitoAppStatus() error = %v, expectErr %v", err, tt.err)
				return
			}
			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("updateKogitoAppStatus() result = %v, expectResult %v", result, tt.result)
				return
			}
		})
	}
}

func TestReconcileKogitoApp_PersistenceEnabledWithInfra(t *testing.T) {
	kogitoApp := createFakeKogitoApp()
	imgs := createFakeImages(kogitoApp.Name, map[string]string{framework.LabelKeyOrgKiePersistenceRequired: "true"})
	kogitoApp.Spec.Infra = v1alpha1.KogitoAppInfra{}
	fakeClient := test.CreateFakeClient([]runtime.Object{kogitoApp}, imgs, nil)
	fakeCache := &cachev1.FakeInformers{}
	// call reconcile object and mock image and build clients
	r := &ReconcileKogitoApp{
		client: fakeClient,
		scheme: meta.GetRegisteredSchema(),
		cache:  fakeCache,
	}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      kogitoApp.Name,
			Namespace: kogitoApp.Namespace,
		},
	}
	// first reconcile to create the objects
	result, err := r.Reconcile(req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Requeue)

	// second reconcile should create infra
	result, err = r.Reconcile(req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Requeue)

	kogitoInfra, created, ready, err := infrastructure.EnsureKogitoInfra(kogitoApp.Namespace, fakeClient).WithInfinispan()
	assert.NoError(t, err)
	assert.False(t, created)      // created in reconciliation phase
	assert.False(t, ready)        // not ready, we don't have status
	assert.NotNil(t, kogitoInfra) // must exist a infra

	dc := &appsv1.DeploymentConfig{ObjectMeta: metav1.ObjectMeta{Name: kogitoApp.Name, Namespace: kogitoApp.Namespace}}
	exists, err := kubernetes.ResourceC(fakeClient).Fetch(dc)
	assert.NoError(t, err)
	assert.True(t, exists) // Already created in the first reconciliation phase

	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: kogitores.GenerateProtoBufConfigMapName(kogitoApp), Namespace: kogitoApp.Namespace}}
	exists, err = kubernetes.ResourceC(fakeClient).Fetch(cm)
	assert.NoError(t, err)
	assert.True(t, exists) // Already created in the first reconciliation phase
	assert.NotNil(t, cm)
}

type mockCache struct {
	cache.Cache
	kogitoApp *v1alpha1.KogitoApp
}

func (c mockCache) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	app := obj.(*v1alpha1.KogitoApp)
	app.Spec = c.kogitoApp.Spec
	app.Status = c.kogitoApp.Status
	return nil
}
