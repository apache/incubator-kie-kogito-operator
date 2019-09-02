package kogitoapp

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"

	cachev1 "sigs.k8s.io/controller-runtime/pkg/cache/informertest"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kogitoclient "github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/builder"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/shared"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	dockerv10 "github.com/openshift/api/image/docker10"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	buildfake "github.com/openshift/client-go/build/clientset/versioned/fake"
	imgfake "github.com/openshift/client-go/image/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	cpuResource    = v1alpha1.ResourceCPU
	memoryResource = v1alpha1.ResourceMemory
	cpuValue       = "1"
	cr             = v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kogito-operator",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Name: "test-app",
			Resources: v1alpha1.Resources{
				Limits: []v1alpha1.ResourceMap{
					{
						Resource: cpuResource,
						Value:    cpuValue,
					},
				},
			},
		},
	}
)

func TestNewContainerWithResource(t *testing.T) {
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
	gitURL := "https://github.com/kiegroup/kogito-examples/"
	kogitoapp := &cr
	kogitoapp.Spec.Build = &v1alpha1.KogitoAppBuildObject{
		GitSource: &v1alpha1.GitSource{
			URI:        &gitURL,
			ContextDir: "jbpm-quarkus-example",
		},
	}
	dockerImageRaw, err := json.Marshal(&dockerv10.DockerImage{
		Config: &dockerv10.DockerConfig{
			Labels: map[string]string{
				openshift.ImageLabelForExposeServices: "8080:http",
			},
		},
	})
	isTag := imgv1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s:latest", kogitoapp.Spec.Name),
			Namespace: "test",
		},
		Image: imgv1.Image{
			DockerImageMetadata: runtime.RawExtension{
				Raw: dockerImageRaw,
			},
		},
	}
	objs := []runtime.Object{kogitoapp}
	buildToTrigger, _ := builder.NewBuildConfigS2I(kogitoapp)
	// add to schemas to avoid: "failed to add object to fake client"
	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion,
		kogitoapp,
		&v1alpha1.KogitoAppList{})
	s.AddKnownTypes(appsv1.SchemeGroupVersion,
		&appsv1.DeploymentConfig{},
		&appsv1.DeploymentConfigList{})
	s.AddKnownTypes(buildv1.SchemeGroupVersion, &buildv1.BuildConfig{})
	s.AddKnownTypes(routev1.SchemeGroupVersion, &routev1.Route{})
	s.AddKnownTypes(imgv1.SchemeGroupVersion, &imgv1.ImageStreamTag{}, &imgv1.ImageStream{})
	// Create a fake client to mock API calls.
	cli := fake.NewFakeClient(objs...)
	// OpenShift Image Client Fake with image tag defined and image built
	imgcli := imgfake.NewSimpleClientset(&isTag).ImageV1()
	// OpenShift Build Client Fake with build for s2i defined, since we'll trigger a build during the reconcile phase
	buildcli := buildfake.NewSimpleClientset(&buildToTrigger).BuildV1()
	// ********** sanity check
	kogitoAppList := &v1alpha1.KogitoAppList{}
	err = cli.List(context.TODO(), &client.ListOptions{Namespace: "test"}, kogitoAppList)
	if err != nil {
		t.Fatalf("Failed to list kogitoapp (%v)", err)
	}
	assert.True(t, len(kogitoAppList.Items) > 0)
	assert.True(t, kogitoAppList.Items[0].Spec.Resources.Limits[0].Resource == cpuResource)
	cache := &cachev1.FakeInformers{}
	// call reconcile object and mock image and build clients
	r := &ReconcileKogitoApp{
		client: &kogitoclient.Client{
			BuildCli:   buildcli,
			ControlCli: cli,
			ImageCli:   imgcli,
		},
		scheme: s,
		cache:  cache,
	}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      kogitoapp.Name,
			Namespace: kogitoapp.Namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}
	// Let's verify if the objects have been built
	dc := &appsv1.DeploymentConfig{}
	_, err = kubernetes.ResourceC(r.client).FetchWithKey(types.NamespacedName{Name: kogitoapp.Spec.Name, Namespace: kogitoapp.Namespace}, dc)
	assert.NoError(t, err)
	assert.NotNil(t, dc)
	assert.Len(t, dc.Spec.Template.Spec.Containers, 1)
	assert.Len(t, dc.GetOwnerReferences(), 1)
}
