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
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestIsOwner(t *testing.T) {
	uuid := test.GenerateUID()
	type args struct {
		resource resource.KubernetesResource
		owner    resource.KubernetesResource
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"I own you",
			args{
				resource: &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "config-map", Namespace: t.Name(),
					OwnerReferences: []metav1.OwnerReference{{
						Kind: "Deployment",
						Name: "deployment",
						UID:  uuid,
					}}},
				},
				owner: &apps.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "deployment", Namespace: t.Name(), UID: uuid}},
			},
			true,
		},
		{
			"I don't own you",
			args{
				resource: &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "config-map", Namespace: t.Name()}},
				owner:    &apps.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "deployment", Namespace: t.Name(), UID: uuid}},
			},
			false,
		},
		{
			"I'm not the only one :(",
			args{
				resource: &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "config-map", Namespace: t.Name(),
					OwnerReferences: []metav1.OwnerReference{
						{
							Kind: "Deployment",
							Name: "deployment",
							UID:  uuid,
						},
						{
							Kind: "BuildConfig",
							Name: "the-builder",
							UID:  test.GenerateUID(),
						},
					}},
				},
				owner: &apps.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "deployment", Namespace: t.Name(), UID: uuid}},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsOwner(tt.args.resource, tt.args.owner)
			if got != tt.want {
				t.Errorf("IsOwner() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddOwnerReference(t *testing.T) {
	scheme := meta.GetRegisteredSchema()
	owner := &apps.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "deployment", Namespace: t.Name(), UID: test.GenerateUID()}}
	owned := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "config-map", Namespace: t.Name(), UID: test.GenerateUID()}}

	err := AddOwnerReference(owner, scheme, owned)
	assert.NoError(t, err)
	assert.Len(t, owned.OwnerReferences, 1)

	err = AddOwnerReference(owner, scheme, owned)
	assert.NoError(t, err)
	assert.Len(t, owned.OwnerReferences, 1)
}

func TestRemoveOwnerReference(t *testing.T) {
	namespace := t.Name()
	scheme := meta.GetRegisteredSchema()

	travels := &v1beta1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "travels",
			Namespace: namespace,
			UID:       test.GenerateUID(),
		},
	}
	visas := &v1beta1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "visas",
			Namespace: namespace,
			UID:       test.GenerateUID(),
		},
	}

	kogitoInfra := &v1beta1.KogitoInfra{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "infinispan-infra",
			Namespace: namespace,
		},
	}
	err := AddOwnerReference(travels, scheme, kogitoInfra)
	assert.NoError(t, err)
	err = AddOwnerReference(visas, scheme, kogitoInfra)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(kogitoInfra.GetOwnerReferences()))
	RemoveOwnerReference(travels, kogitoInfra)
	assert.Equal(t, 1, len(kogitoInfra.GetOwnerReferences()))
	ownerReference := kogitoInfra.GetOwnerReferences()[0]
	assert.Equal(t, visas.UID, ownerReference.UID)
	assert.Equal(t, visas.Name, ownerReference.Name)
}
