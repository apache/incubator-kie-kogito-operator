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

package test

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"

	buildfake "github.com/openshift/client-go/build/clientset/versioned/fake"
	imgfake "github.com/openshift/client-go/image/clientset/versioned/fake"
)

// CreateFakeClient will create a fake client for mock test
func CreateFakeClient(objects []runtime.Object, imageObjs []runtime.Object, buildObjs []runtime.Object) (*client.Client, *runtime.Scheme) {
	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion,
		&v1alpha1.KogitoApp{},
		&v1alpha1.KogitoAppList{},
		&v1alpha1.KogitoDataIndex{},
		&v1alpha1.KogitoDataIndexList{},
	)
	s.AddKnownTypes(appsv1.SchemeGroupVersion,
		&appsv1.DeploymentConfig{},
		&appsv1.DeploymentConfigList{})
	s.AddKnownTypes(buildv1.SchemeGroupVersion, &buildv1.BuildConfig{})
	s.AddKnownTypes(routev1.SchemeGroupVersion, &routev1.Route{})
	s.AddKnownTypes(imgv1.SchemeGroupVersion, &imgv1.ImageStreamTag{}, &imgv1.ImageStream{})
	// Create a fake client to mock API calls.
	cli := fake.NewFakeClient(objects...)
	// OpenShift Image Client Fake with image tag defined and image built
	imgcli := imgfake.NewSimpleClientset(imageObjs...).ImageV1()
	// OpenShift Build Client Fake with build for s2i defined, since we'll trigger a build during the reconcile phase
	buildcli := buildfake.NewSimpleClientset(buildObjs...).BuildV1()

	return &client.Client{
		ControlCli: cli,
		BuildCli:   buildcli,
		ImageCli:   imgcli,
	}, s
}
