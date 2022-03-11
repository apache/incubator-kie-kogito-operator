package shared

import (
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-operator/core/kogitoservice"
	"github.com/kiegroup/kogito-operator/core/operator"
	"github.com/kiegroup/kogito-operator/core/test"
	"github.com/kiegroup/kogito-operator/meta"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"testing"
)

func TestCreateRequiredResources(t *testing.T) {
	ns := t.Name()
	runtimeService := test.CreateFakeKogitoRuntime(ns)
	runtimeDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runtimeService.Name,
			Namespace: runtimeService.Namespace,
		},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: 1,
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(runtimeDeployment).Build()
	context := operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}

	protoBufFilePath := "/persistence/protobuf/list.json"
	responseWithProtoBufData := `[]`
	server := test.MockKogitoSvcReplies(t, test.ServerHandler{Path: protoBufFilePath, JSONResponse: responseWithProtoBufData})
	defer server.Close()
	err := os.Setenv(kogitoservice.envVarKogitoServiceURL, server.URL)
	assert.NoError(t, err)
	protobufConfigMapReconciler := NewProtoBufConfigMapReconciler(context, runtimeService)
	err = protobufConfigMapReconciler.Reconcile()
	assert.NoError(t, err)

	configMaps := &v1.ConfigMapList{}
	err = kubernetes.ResourceC(cli).ListWithNamespace(ns, configMaps)
	assert.NoError(t, err)
	assert.Empty(t, configMaps.Items)
}
