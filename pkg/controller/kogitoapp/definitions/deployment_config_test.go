package definitions

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	dockerv10 "github.com/openshift/api/image/docker10"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_deploymentConfigResource_NewWithValidDocker(t *testing.T) {
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
				ImageLabelForExposeServices:          "8080:http,8181;https",
				orgKieNamespaceLabelKey + "operator": "kogito",
			},
		},
	}
	bcS2I, _ := NewBuildConfigS2I(kogitoApp)
	bcSvc, _ := NewBuildConfigService(kogitoApp, &bcS2I)
	sa := NewServiceAccount(kogitoApp)
	dc, err := NewDeploymentConfig(kogitoApp, &bcSvc, &sa, dockerImage)
	assert.Nil(t, err)
	assert.NotNil(t, dc)
	// we should have only one port. the 8181 is invalid.
	assert.Len(t, dc.Spec.Template.Spec.Containers[0].Ports, 1)
	assert.Equal(t, int32(8080), dc.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort)
	// this one where added by the docker image :)
	assert.Equal(t, "kogito", dc.Labels["operator"])
}

func Test_deploymentConfigResource_NewWithInvalidDocker(t *testing.T) {
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
	bcS2I, _ := NewBuildConfigS2I(kogitoApp)
	bcSvc, _ := NewBuildConfigService(kogitoApp, &bcS2I)
	sa := NewServiceAccount(kogitoApp)
	dc, err := NewDeploymentConfig(kogitoApp, &bcSvc, &sa, &dockerv10.DockerImage{})
	assert.Nil(t, err)
	assert.NotNil(t, dc)
	assert.Len(t, dc.Spec.Selector, 1)
	assert.Len(t, dc.Spec.Template.Spec.Containers, 1)
	assert.Equal(t, bcSvc.Spec.Output.To.Name, dc.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, sa.Name, dc.Spec.Template.Spec.ServiceAccountName)
	assert.Equal(t, "test", dc.Labels[LabelKeyAppName])
	assert.Equal(t, "test", dc.Spec.Selector[LabelKeyAppName])
	assert.Equal(t, "test", dc.Spec.Template.Labels[LabelKeyAppName])
}
