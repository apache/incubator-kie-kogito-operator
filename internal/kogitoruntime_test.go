package internal

import (
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	test2 "github.com/kiegroup/kogito-cloud-operator/core/test"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestFetchKogitoRuntimeService_InstanceFound(t *testing.T) {
	ns := t.Name()
	name := "kogito-runtime"
	kogitoRuntime := &v1beta1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
	cli := test2.NewFakeClientBuilder().AddK8sObjects(kogitoRuntime).Build()
	runtimeHandler := NewKogitoRuntimeHandler(cli, logger.GetLogger("KogitoRuntime"))
	instance, err := runtimeHandler.FetchKogitoRuntimeInstance(types.NamespacedName{Name: name, Namespace: ns})
	assert.NoError(t, err)
	assert.NotNil(t, instance)
}

func TestFetchKogitoRuntimeService_InstanceNotFound(t *testing.T) {
	ns := t.Name()
	name := "kogito-runtime"
	cli := test2.NewFakeClientBuilder().Build()
	runtimeHandler := NewKogitoRuntimeHandler(cli, logger.GetLogger("KogitoRuntime"))
	instance, err := runtimeHandler.FetchKogitoRuntimeInstance(types.NamespacedName{Name: name, Namespace: ns})
	assert.NoError(t, err)
	assert.Nil(t, instance)
}
