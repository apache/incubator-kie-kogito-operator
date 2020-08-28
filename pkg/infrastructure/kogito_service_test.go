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

package infrastructure

import (
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/api/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_getKogitoDataIndexRoute(t *testing.T) {
	ns := t.Name()
	expectedRoute := "http://dataindex-route.com"
	dataIndexes := &v1alpha1.KogitoDataIndexList{
		Items: []v1alpha1.KogitoDataIndex{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DefaultDataIndexName,
					Namespace: ns,
				},
				Status: v1alpha1.KogitoDataIndexStatus{
					KogitoServiceStatus: v1alpha1.KogitoServiceStatus{ExternalURI: expectedRoute},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      DefaultDataIndexName + "2",
					Namespace: ns,
				},
				Status: v1alpha1.KogitoDataIndexStatus{
					KogitoServiceStatus: v1alpha1.KogitoServiceStatus{ExternalURI: ""},
				},
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{dataIndexes}, nil, nil)

	route, err := getSingletonKogitoServiceRoute(cli, ns, &v1alpha1.KogitoDataIndexList{})
	assert.NoError(t, err)
	assert.Equal(t, expectedRoute, route)
}

func Test_getKogitoDataIndexRoute_NoDataIndex(t *testing.T) {
	cli := test.CreateFakeClient(nil, nil, nil)
	route, err := getSingletonKogitoServiceRoute(cli, t.Name(), &v1alpha1.KogitoDataIndexList{})
	assert.NoError(t, err)
	assert.Empty(t, route)
}
