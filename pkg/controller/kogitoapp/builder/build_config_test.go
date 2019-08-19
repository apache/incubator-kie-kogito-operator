package builder

import (
	"testing"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_buildConfigResource_New(t *testing.T) {
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
	bcS2I, err := NewBuildConfigS2I(kogitoApp)
	assert.Nil(t, err)
	assert.NotNil(t, bcS2I)
	bcService, err := NewBuildConfigService(kogitoApp, &bcS2I)
	assert.Nil(t, err)
	assert.NotNil(t, bcService)

	assert.Contains(t, bcS2I.Spec.Output.To.Name, nameSuffix)
	assert.NotContains(t, bcService.Spec.Output.To.Name, nameSuffix)
	assert.Len(t, bcService.Spec.Triggers, 1)
	assert.Len(t, bcS2I.Spec.Triggers, 0)
	assert.Equal(t, bcService.Spec.Source.Images[0].From, *bcS2I.Spec.Output.To)
}
