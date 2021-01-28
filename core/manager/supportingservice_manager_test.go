package manager

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	test2 "github.com/kiegroup/kogito-cloud-operator/core/test"
	api2 "github.com/kiegroup/kogito-cloud-operator/core/test/api"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_ensureSingletonService(t *testing.T) {
	ns := t.Name()
	instance1 := &api2.KogitoSupportingServiceTest{
		ObjectMeta: v1.ObjectMeta{
			Name:      "data-index1",
			Namespace: ns,
		},
		Spec: api2.KogitoSupportingServiceSpecTest{
			ServiceType: api.DataIndex,
		},
	}
	instance2 := &api2.KogitoSupportingServiceTest{
		ObjectMeta: v1.ObjectMeta{
			Name:      "data-index2",
			Namespace: ns,
		},
		Spec: api2.KogitoSupportingServiceSpecTest{
			ServiceType: api.DataIndex,
		},
	}

	cli := test2.NewFakeClientBuilder().AddK8sObjects(instance1, instance2).OnOpenShift().Build()
	supportingServiceHandler := test2.CreateFakeKogitoSupportingServiceHandler(cli)
	supportingServiceManager := NewKogitoSupportingServiceManager(cli, test2.TestLogger, supportingServiceHandler)
	assert.Errorf(t, supportingServiceManager.EnsureSingletonService(ns, api.DataIndex), "kogito Supporting Service(%s) already exists, please delete the duplicate before proceeding", api.DataIndex)
}
