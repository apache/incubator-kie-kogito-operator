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

package status

import (
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitodataindex/resource"
	commonres "github.com/kiegroup/kogito-cloud-operator/pkg/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_ManageStatus_WhenTheresStatusChange(t *testing.T) {
	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			Name:     "my-data-index",
			Replicas: 1,
		},
	}
	client, _ := test.CreateFakeClient([]runtime.Object{instance}, nil, nil)
	resources, err := resource.CreateOrFetchResources(instance, commonres.FactoryContext{Client: client})
	assert.NoError(t, err)

	err = ManageStatus(instance, &resources, client)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.NotNil(t, instance.Status.Conditions)
	assert.Len(t, instance.Status.Conditions, 1)
}
