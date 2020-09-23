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

package kogitotrustyui

import (
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	imagev1 "github.com/openshift/api/image/v1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestReconcileKogitoTrustyUI_Reconcile(t *testing.T) {
	replicas := int32(1)
	instance := &v1alpha1.KogitoTrustyUI{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultTrustyUIName, Namespace: t.Name()},
		Spec: v1alpha1.KogitoTrustyUISpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance}, nil, nil)

	r := ReconcileKogitoTrustyUI{client: cli, scheme: meta.GetRegisteredSchema()}

	// first reconciliation
	test.AssertReconcileMustNotRequeue(t, &r, instance)
	// second time
	test.AssertReconcileMustNotRequeue(t, &r, instance)

	_, err := kubernetes.ResourceC(cli).Fetch(instance)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.Len(t, instance.Status.Conditions, 1)
}

// see: https://issues.redhat.com/browse/KOGITO-2535
func TestReconcileKogitoTrustyUI_CustomImage(t *testing.T) {
	replicas := int32(1)
	instance := &v1alpha1.KogitoTrustyUI{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultTrustyUIName, Namespace: t.Name()},
		Spec: v1alpha1.KogitoTrustyUISpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Replicas: &replicas,
				Image:    "quay.io/mynamespace/awesome-trusty-ui:0.1.3",
			},
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance}, nil, nil)
	test.AssertReconcileMustNotRequeue(t, &ReconcileKogitoTrustyUI{client: cli, scheme: meta.GetRegisteredSchema()}, instance)
	// image stream
	is := imagev1.ImageStream{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultTrustyUIImageName, Namespace: instance.Namespace},
	}
	exists, err := kubernetes.ResourceC(cli).Fetch(&is)
	assert.True(t, exists)
	assert.NoError(t, err)
	assert.Len(t, is.Spec.Tags, 1)
	assert.Equal(t, "0.1.3", is.Spec.Tags[0].Name)
	assert.Equal(t, "quay.io/mynamespace/awesome-trusty-ui:0.1.3", is.Spec.Tags[0].From.Name)

}
