package kogitoservice

import (
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/core/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAddFinalizer(t *testing.T) {
	ns := t.Name()
	dataIndex := test.CreateFakeDataIndex(ns)
	cli := test.NewFakeClientBuilder().AddK8sObjects(dataIndex).Build()
	infraHandler := test.CreateFakeKogitoInfraHandler(cli)
	finalizerHandler := NewFinalizerHandler(cli, logger.GetLogger("finalizer"), meta.GetRegisteredSchema(), infraHandler)
	err := finalizerHandler.AddFinalizer(dataIndex)
	assert.NoError(t, err)
	exists, err := kubernetes.ResourceC(cli).Fetch(dataIndex)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, 1, len(dataIndex.GetFinalizers()))
}

func TestHandleFinalization(t *testing.T) {
	ns := t.Name()
	dataIndex := test.CreateFakeDataIndex(ns)
	dataIndex.SetFinalizers([]string{"delete.kogitoInfra.ownership.finalizer"})
	cli := test.NewFakeClientBuilder().AddK8sObjects(dataIndex).Build()
	infraHandler := test.CreateFakeKogitoInfraHandler(cli)
	finalizerHandler := NewFinalizerHandler(cli, logger.GetLogger("finalizer"), meta.GetRegisteredSchema(), infraHandler)
	err := finalizerHandler.HandleFinalization(dataIndex)
	assert.NoError(t, err)
	exists, err := kubernetes.ResourceC(cli).Fetch(dataIndex)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, 0, len(dataIndex.GetFinalizers()))
}
