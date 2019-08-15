package inventory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_CreateNamespaceThatDoesNotExist(t *testing.T) {
	cli := fake.NewFakeClient()
	ns, err := NamespaceC(&Client{Cli: cli}).CreateIfNotExists("test")
	assert.Nil(t, err)
	assert.NotNil(t, ns)
}

func Test_FetchNamespaceThatDoesNotExist(t *testing.T) {
	cli := fake.NewFakeClient()
	ns, err := NamespaceC(&Client{Cli: cli}).Fetch("test")
	assert.Nil(t, err)
	assert.Nil(t, ns)
}

func Test_FetchNamespaceThatDExists(t *testing.T) {
	cli := fake.NewFakeClient(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test", CreationTimestamp: metav1.Now()}})
	ns, err := NamespaceC(&Client{Cli: cli}).Fetch("test")
	assert.Nil(t, err)
	assert.NotNil(t, ns)
	assert.False(t, ns.CreationTimestamp.IsZero())
}
