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

package kogitosupportingservice

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestReconcileKogitoSupportingService_Reconcile(t *testing.T) {
	replicas := int32(1)
	instance := &v1alpha1.KogitoSupportingService{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultJobsServiceName, Namespace: t.Name()},
		Spec: v1alpha1.KogitoSupportingServiceSpec{
			ServiceType:       v1alpha1.JobsService,
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects([]runtime.Object{instance}).OnOpenShift().Build()

	r := &JobsServiceSupportingServiceResource{}
	// first reconciliation

	requeue, err := r.Reconcile(cli, instance, meta.GetRegisteredSchema())
	assert.NoError(t, err)
	assert.False(t, requeue)
	// second time
	requeue, err = r.Reconcile(cli, instance, meta.GetRegisteredSchema())
	assert.NoError(t, err)
	assert.False(t, requeue)

	_, err = kubernetes.ResourceC(cli).Fetch(instance)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.Len(t, instance.Status.Conditions, 1)
}

func Test_isKogitoInfraUpdated(t *testing.T) {
	oldKogitoInfra := v1alpha1.KogitoInfra{
		Status: v1alpha1.KogitoInfraStatus{
			AppProps: map[string]string{
				"myprops": "custom-value",
			},
			Env: []v12.EnvVar{
				{
					Name:  "myenv",
					Value: "custom-value",
				},
			},
		},
	}

	newKogitoInfra := v1alpha1.KogitoInfra{
		Status: v1alpha1.KogitoInfraStatus{
			AppProps: map[string]string{
				"myprops": "custom-value",
			},
			Env: []v12.EnvVar{
				{
					Name:  "myenv",
					Value: "custom-value",
				},
			},
		},
	}
	// No change test
	assert.False(t, isKogitoInfraUpdated(&oldKogitoInfra, &newKogitoInfra))

	// Infra updated with some new AppProps
	newKogitoInfra = v1alpha1.KogitoInfra{
		Status: v1alpha1.KogitoInfraStatus{
			AppProps: map[string]string{
				"myprops": "custom-value",
				"newprop": "new-custom-value",
			},
			Env: []v12.EnvVar{
				{
					Name:  "myenv",
					Value: "custom-value",
				},
			},
		},
	}
	// AppProps changed
	assert.True(t, isKogitoInfraUpdated(&oldKogitoInfra, &newKogitoInfra))

	// new env added
	newKogitoInfra = v1alpha1.KogitoInfra{
		Status: v1alpha1.KogitoInfraStatus{
			AppProps: map[string]string{
				"myprops": "custom-value",
			},
			Env: []v12.EnvVar{
				{
					Name:  "myenv",
					Value: "custom-value",
				},
				{
					Name:  "new-env",
					Value: "new-custom-value",
				},
			},
		},
	}

	// Env Changed
	assert.True(t, isKogitoInfraUpdated(&oldKogitoInfra, &newKogitoInfra))
}

func TestContains(t *testing.T) {
	allServices := []v1alpha1.ServiceType{
		v1alpha1.MgmtConsole,
		v1alpha1.JobsService,
		v1alpha1.TrustyAI,
	}
	testService := v1alpha1.DataIndex

	assert.False(t, contains(allServices, testService))
}

func Test_ensureSingletonService(t *testing.T) {
	ns := t.Name()
	instance1 := &v1alpha1.KogitoSupportingService{
		ObjectMeta: v1.ObjectMeta{
			Name:      "data-index1",
			Namespace: ns,
		},
		Spec: v1alpha1.KogitoSupportingServiceSpec{
			ServiceType: v1alpha1.DataIndex,
		},
	}
	instance2 := &v1alpha1.KogitoSupportingService{
		ObjectMeta: v1.ObjectMeta{
			Name:      "data-index2",
			Namespace: ns,
		},
		Spec: v1alpha1.KogitoSupportingServiceSpec{
			ServiceType: v1alpha1.DataIndex,
		},
	}

	cli := test.NewFakeClientBuilder().AddK8sObjects([]runtime.Object{instance1, instance2}).OnOpenShift().Build()
	assert.Errorf(t, ensureSingletonService(cli, ns, v1alpha1.DataIndex), "kogito Supporting Service(%s) already exists, please delete the duplicate before proceeding", v1alpha1.DataIndex)

}
