package inventory

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/definitions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var mockKogitoApp = &v1alpha1.KogitoApp{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test",
		Namespace: "test-ns",
	},
	Spec: v1alpha1.KogitoAppSpec{
		Name: "example-app",
	},
}

func TestCreateIfNotExists_WhenResourceNotExists(t *testing.T) {
	sa := definitions.NewServiceAccount(mockKogitoApp)
	created, err := CreateResourceIfNotExists(fake.NewFakeClient(), &sa)
	assert.Nil(t, err)
	assert.True(t, created)
}

func TestCreateIfNotExists_WhenResourceExists(t *testing.T) {
	sa := definitions.NewServiceAccount(mockKogitoApp)
	// here we tell to the fake client that this object already exists in the cluster
	created, err := CreateResourceIfNotExists(fake.NewFakeClient([]runtime.Object{&sa}...), &sa)
	assert.Nil(t, err)
	assert.False(t, created)
}
