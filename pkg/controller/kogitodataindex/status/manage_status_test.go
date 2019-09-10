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
