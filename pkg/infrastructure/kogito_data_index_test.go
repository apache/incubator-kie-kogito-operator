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

package infrastructure

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func Test_getKogitoDataIndexRoute(t *testing.T) {
	ns := t.Name()
	expectedRoute := "http://dataindex-route.com"
	dataIndexes := &v1alpha1.KogitoDataIndexList{
		Items: []v1alpha1.KogitoDataIndex{
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "kogito-data-index",
					Namespace: ns,
				},
				Status: v1alpha1.KogitoDataIndexStatus{
					Route: expectedRoute,
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "kogito-data-index2",
					Namespace: ns,
				},
				Status: v1alpha1.KogitoDataIndexStatus{
					Route: "",
				},
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{dataIndexes}, nil, nil)

	route, err := getKogitoDataIndexRoute(cli, ns)
	assert.NoError(t, err)
	assert.Equal(t, expectedRoute, route)
}

func Test_getKogitoDataIndexRoute_NoDataIndex(t *testing.T) {
	cli := test.CreateFakeClient(nil, nil, nil)
	route, err := getKogitoDataIndexRoute(cli, t.Name())
	assert.NoError(t, err)
	assert.Empty(t, route)
}

func TestInjectDataIndexURLIntoKogitoApps(t *testing.T) {
	ns := t.Name()
	name := "my-kogito-app"
	expectedRoute := "http://dataindex-route.com"
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
	dataIndexes := &v1alpha1.KogitoDataIndexList{
		Items: []v1alpha1.KogitoDataIndex{
			{
				ObjectMeta: v1.ObjectMeta{
					Name:      "kogito-data-index",
					Namespace: ns,
				},
				Status: v1alpha1.KogitoDataIndexStatus{
					Route: expectedRoute,
				},
			},
		},
	}
	client := test.CreateFakeClient([]runtime.Object{kogitoApp, dataIndexes}, nil, nil)

	err := InjectDataIndexURLIntoKogitoApps(client, ns)
	assert.NoError(t, err)

	exist, err := kubernetes.ResourceC(client).Fetch(kogitoApp)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Contains(t, kogitoApp.Spec.Env, v1alpha1.Env{Name: kogitoDataIndexRouteEnv, Value: expectedRoute})
}
