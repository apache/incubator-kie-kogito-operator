package internal

import (
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	test2 "github.com/kiegroup/kogito-cloud-operator/core/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestFetchKogitoInfraInstance_InstanceFound(t *testing.T) {
	ns := t.Name()
	name := "InfinispanInfra"
	kogitoInfra := &v1beta1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
	cli := test2.NewFakeClientBuilder().AddK8sObjects(kogitoInfra).Build()
	infraHandler := NewKogitoInfraHandler(cli, logger.GetLogger("KogitoInfra"))
	instance, err := infraHandler.FetchKogitoInfraInstance(types.NamespacedName{Name: name, Namespace: ns})
	assert.NoError(t, err)
	assert.NotNil(t, instance)
}

func TestFetchKogitoInfraInstance_InstanceNotFound(t *testing.T) {
	ns := t.Name()
	name := "InfinispanInfra"
	cli := test2.NewFakeClientBuilder().Build()
	infraHandler := NewKogitoInfraHandler(cli, logger.GetLogger("KogitoInfra"))
	instance, err := infraHandler.FetchKogitoInfraInstance(types.NamespacedName{Name: name, Namespace: ns})
	assert.NoError(t, err)
	assert.Nil(t, instance)
}
