package kogitoapp

import (
	"context"
	"testing"

	resource "k8s.io/apimachinery/pkg/api/resource"

	corev1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	cpuResource    = v1alpha1.ResourceCPU
	memoryResource = v1alpha1.ResourceMemory
)

func TestNewContainerWithResource(t *testing.T) {
	cpuValue := "1"
	cr := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Resources: v1alpha1.Resources{
				Limits: []v1alpha1.ResourceMap{
					{
						Resource: &cpuResource,
						Value:    &cpuValue,
					},
				},
			},
		},
	}
	container := corev1.Container{
		Name:            cr.Spec.Name,
		Env:             shared.FromEnvToEnvVar(cr.Spec.Env),
		Resources:       *shared.FromResourcesToResourcesRequirements(cr.Spec.Resources),
		ImagePullPolicy: corev1.PullAlways,
	}
	assert.NotNil(t, container)
	cpuQty := resource.MustParse(cpuValue)
	assert.Equal(t, container.Resources.Limits.Cpu(), &cpuQty)
	assert.Equal(t, container.Resources.Requests.Cpu(), &resource.Quantity{Format: resource.DecimalSI})
}

func TestKogitoAppWithResource(t *testing.T) {
	valueCPU := "1"
	gitURL := "https://github.com/kiegroup/kogito-examples/"
	kogitoBuild := v1alpha1.KogitoAppBuildObject{
		GitSource: &v1alpha1.GitSource{
			URI:        &gitURL,
			ContextDir: "jbpm-quarkus-example",
		},
	}
	kogitoapp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Build: &kogitoBuild,
			Resources: v1alpha1.Resources{
				Limits: []v1alpha1.ResourceMap{
					{
						Resource: &cpuResource,
						Value:    &valueCPU,
					},
				},
			},
		},
	}
	objs := []runtime.Object{kogitoapp}
	// add to schemas to avoid: "failed to add object to fake client"
	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, kogitoapp)
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion, &v1alpha1.KogitoAppList{})
	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	// ********** sanity check
	kogitoAppList := &v1alpha1.KogitoAppList{}
	err := cl.List(context.TODO(), &client.ListOptions{Namespace: "test"}, kogitoAppList)
	if err != nil {
		t.Fatalf("Failed to list kogitoapp (%v)", err)
	}
	assert.True(t, len(kogitoAppList.Items) > 0)
	assert.True(t, *kogitoAppList.Items[0].Spec.Resources.Limits[0].Resource == cpuResource)

	// TODO: call reconcile object and mock image and build clients
}
