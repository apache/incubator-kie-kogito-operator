package kogitodataindex

import (
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
)

func TestReconcileKogitoDataIndex_Reconcile(t *testing.T) {
	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			Name: "my-data-index",
		},
	}
	client, s := test.CreateFakeClient([]runtime.Object{instance}, nil, nil)
	r := &ReconcileKogitoDataIndex{
		client: client,
		scheme: s,
	}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
}
