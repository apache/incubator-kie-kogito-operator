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

package framework

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	obuildv1 "github.com/openshift/api/build/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func Test_controllerWatcher_WatchWithOCPObjectsOnKubernetes(t *testing.T) {
	cli := test.CreateFakeClient(nil, nil, nil)
	controller := test.NewController()
	manager := test.NewManager()
	requiredObjects := []WatchedObjects{
		{
			GroupVersion: obuildv1.GroupVersion,
			AddToScheme:  obuildv1.Install,
			Objects:      []runtime.Object{&obuildv1.BuildConfig{}},
		},
		{
			Objects: []runtime.Object{&corev1.Service{}, &corev1.ConfigMap{}},
		},
	}

	watcher := NewControllerWatcher(cli, manager, controller, &v1alpha1.KogitoApp{})
	assert.NotNil(t, watcher)

	err := watcher.Watch(requiredObjects...)
	assert.NoError(t, err)
	// we are not on OpenShift
	assert.False(t, watcher.AreAllObjectsWatched())
	// we should only have the objects from Kubernetes core
	assert.Len(t, controller.GetWatchedSources(), 2)
}

func Test_controllerWatcher_WatchWithOCPObjectsOnOpenShift(t *testing.T) {
	cli := test.CreateFakeClientOnOpenShift(nil, nil, nil)
	controller := test.NewController()
	manager := test.NewManager()
	requiredObjects := []WatchedObjects{
		{
			GroupVersion: obuildv1.GroupVersion,
			AddToScheme:  obuildv1.Install,
			Objects:      []runtime.Object{&obuildv1.BuildConfig{}},
		},
		{
			Objects: []runtime.Object{&corev1.Service{}, &corev1.ConfigMap{}},
		},
	}

	watcher := NewControllerWatcher(cli, manager, controller, &v1alpha1.KogitoApp{})
	assert.NotNil(t, watcher)

	err := watcher.Watch(requiredObjects...)
	assert.NoError(t, err)
	// we are on OpenShift
	assert.True(t, watcher.AreAllObjectsWatched())
	// we should only have the objects from Kubernetes core
	assert.Len(t, controller.GetWatchedSources(), 3)
}
