// Copyright 2021 Red Hat, Inc. and/or its affiliates
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
	"github.com/kiegroup/kogito-operator/api/v1beta1"
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/test"
	"github.com/kiegroup/kogito-operator/meta"
	"github.com/kiegroup/kogito-operator/version"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestReconcileKogitoRuntimeFinalizer_AddFinalizer_Kubernetes(t *testing.T) {
	ns := t.Name()
	instance := test.CreateFakeKogitoRuntime(ns)

	cli := test.NewFakeClientBuilder().AddK8sObjects(instance).Build()
	r := FinalizeKogitoRuntime{Client: cli, Scheme: meta.GetRegisteredSchema()}
	test.AssertReconcileMustNotRequeue(t, &r, instance)

	updatedInstance := &v1beta1.KogitoRuntime{ObjectMeta: v1.ObjectMeta{
		Name:      instance.GetName(),
		Namespace: instance.GetNamespace(),
	}}
	_, err := kubernetes.ResourceC(cli).Fetch(updatedInstance)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(updatedInstance.GetFinalizers()))
}

func TestReconcileKogitoRuntimeFinalizer_AddFinalizer_Openshift(t *testing.T) {
	ns := t.Name()
	instance := test.CreateFakeKogitoRuntime(ns)

	cli := test.NewFakeClientBuilder().OnOpenShift().AddK8sObjects(instance).Build()
	r := FinalizeKogitoRuntime{Client: cli, Scheme: meta.GetRegisteredSchema()}
	test.AssertReconcileMustNotRequeue(t, &r, instance)

	updatedInstance := &v1beta1.KogitoRuntime{ObjectMeta: v1.ObjectMeta{
		Name:      instance.GetName(),
		Namespace: instance.GetNamespace(),
	}}
	_, err := kubernetes.ResourceC(cli).Fetch(updatedInstance)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(updatedInstance.GetFinalizers()))
}

func TestReconcileKogitoRuntimeFinalizer_RemoveFinalizer_kubernetes(t *testing.T) {
	ns := t.Name()
	instance := test.CreateFakeKogitoRuntime(ns)
	currentTime := v1.Now()
	instance.SetDeletionTimestamp(&currentTime)
	instance.SetFinalizers([]string{"delete.kogitoInfra.ownership.finalizer"})

	cli := test.NewFakeClientBuilder().AddK8sObjects(instance).Build()
	r := FinalizeKogitoRuntime{Client: cli, Scheme: meta.GetRegisteredSchema()}
	test.AssertReconcileMustNotRequeue(t, &r, instance)

	updatedInstance := &v1beta1.KogitoRuntime{ObjectMeta: v1.ObjectMeta{
		Name:      instance.GetName(),
		Namespace: instance.GetNamespace(),
	}}
	_, err := kubernetes.ResourceC(cli).Fetch(updatedInstance)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(updatedInstance.GetFinalizers()))
}

func TestReconcileKogitoRuntimeFinalizer_RemoveFinalizer_Openshift(t *testing.T) {
	ns := t.Name()
	instance := test.CreateFakeKogitoRuntime(ns)
	currentTime := v1.Now()
	instance.SetDeletionTimestamp(&currentTime)
	instance.SetFinalizers([]string{"delete.kogitoInfra.ownership.finalizer", "delete.imageStream.ownership.finalizer"})

	is, ist := test.CreateFakeImageStreams(instance.Name, instance.Namespace, infrastructure.GetKogitoImageVersion(version.Version))

	cli := test.NewFakeClientBuilder().OnOpenShift().AddK8sObjects(instance, is).AddImageObjects(ist).Build()
	r := FinalizeKogitoRuntime{Client: cli, Scheme: meta.GetRegisteredSchema()}
	test.AssertReconcileMustNotRequeue(t, &r, instance)

	updatedInstance := &v1beta1.KogitoRuntime{ObjectMeta: v1.ObjectMeta{
		Name:      instance.GetName(),
		Namespace: instance.GetNamespace(),
	}}
	_, err := kubernetes.ResourceC(cli).Fetch(updatedInstance)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(updatedInstance.GetFinalizers()))
}
