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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	discfake "k8s.io/client-go/discovery/fake"
	clienttesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	monfake "github.com/coreos/prometheus-operator/pkg/client/versioned/fake"
	buildfake "github.com/openshift/client-go/build/clientset/versioned/fake"
	imgfake "github.com/openshift/client-go/image/clientset/versioned/fake"
)

// CreateFakeClient will create a fake client for mock test on Kubernetes env, use cases that depends on OpenShift should use CreateFakeClientOnOpenShift
func CreateFakeClient(objects []runtime.Object, imageObjs []runtime.Object, buildObjs []runtime.Object) *client.Client {
	return CreateFakeClientWithDisco(objects, imageObjs, buildObjs, false)
}

// CreateFakeClientOnOpenShift same as CreateFakeClientWithDisco setting openshift flag to true
func CreateFakeClientOnOpenShift(objects []runtime.Object, imageObjs []runtime.Object, buildObjs []runtime.Object) *client.Client {
	return CreateFakeClientWithDisco(objects, imageObjs, buildObjs, true)
}

// CreateFakeClientWithDisco controls if the Discovery API should return true for `IsOpenShift` calls
func CreateFakeClientWithDisco(objects []runtime.Object, imageObjs []runtime.Object, buildObjs []runtime.Object, openshift bool) *client.Client {
	// Create a fake client to mock API calls.
	cli := fake.NewFakeClientWithScheme(meta.GetRegisteredSchema(), objects...)
	// OpenShift Image Client Fake with image tag defined and image built
	imgcli := imgfake.NewSimpleClientset(imageObjs...).ImageV1()
	// OpenShift Build Client Fake with build for s2i defined, since we'll trigger a build during the reconcile phase
	buildcli := buildfake.NewSimpleClientset(buildObjs...).BuildV1()

	return &client.Client{
		ControlCli:    cli,
		BuildCli:      buildcli,
		ImageCli:      imgcli,
		PrometheusCli: monfake.NewSimpleClientset().MonitoringV1(),
		Discovery:     CreateFakeDiscoveryClient(openshift),
	}
}

// CreateFakeDiscoveryClient creates a fake discovery client that supports prometheus, infinispan, strimzi api
func CreateFakeDiscoveryClient(openshift bool) discovery.DiscoveryInterface {
	disco := &discfake.FakeDiscovery{
		Fake: &clienttesting.Fake{
			Resources: []*metav1.APIResourceList{
				{
					GroupVersion: "monitoring.coreos.com/v1alpha1",
				},
				{
					GroupVersion: "infinispan.org/v1",
				},
				{
					GroupVersion: "kafka.strimzi.io/v1beta1",
				},
			},
		},
	}
	if openshift {
		disco.Fake.Resources = append(disco.Fake.Resources, &metav1.APIResourceList{GroupVersion: "openshift.io/v1"})
	}
	return disco
}
