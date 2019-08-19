package builder

import (
	"testing"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	dockerv10 "github.com/openshift/api/image/docker10"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_serviceResource_NewWithAndWithoutDockerImg(t *testing.T) {
	uri := "https://github.com/kiegroup/kogito-examples"
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Name: "test",
			Build: &v1alpha1.KogitoAppBuildObject{
				GitSource: &v1alpha1.GitSource{
					URI:        &uri,
					ContextDir: "jbpm-quarkus-example",
				},
			},
		},
	}
	dockerImage := &dockerv10.DockerImage{
		Config: &dockerv10.DockerConfig{
			Labels: map[string]string{
				// notice the semicolon
				openshift.ImageLabelForExposeServices: "8080:http",
				orgKieNamespaceLabelKey + "operator":  "kogito",
			},
		},
	}
	bcS2I, _ := NewBuildConfigS2I(kogitoApp)
	bcSvc, _ := NewBuildConfigService(kogitoApp, &bcS2I)
	sa := NewServiceAccount(kogitoApp)
	dc, _ := NewDeploymentConfig(kogitoApp, &bcSvc, &sa, nil)
	svc := NewService(kogitoApp, dc)
	assert.Nil(t, svc)
	// try again, now with ports
	dc, _ = NewDeploymentConfig(kogitoApp, &bcSvc, &sa, dockerImage)
	svc = NewService(kogitoApp, dc)
	assert.NotNil(t, svc)
	assert.Len(t, svc.Spec.Ports, 1)
	assert.Equal(t, int32(8080), svc.Spec.Ports[0].Port)
}
