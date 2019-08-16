package definitions

import (
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var kogitoApp = &v1alpha1.KogitoApp{
	Spec: v1alpha1.KogitoAppSpec{
		Name: "test",
	},
}

func Test_addDefaultMeta_whenLabelsAreNotDefined(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{}
	addDefaultMeta(objectMeta, kogitoApp)
	assert.True(t, objectMeta.Labels[LabelKeyAppName] == "test")
}

func Test_addDefaultMeta_whenAlreadyHasLabels(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{
		Labels: map[string]string{
			"app":      "test123",
			"operator": "kogito",
		},
	}
	addDefaultMeta(objectMeta, kogitoApp)
	assert.True(t, objectMeta.Labels[LabelKeyAppName] == "test")
	assert.True(t, objectMeta.Labels["operator"] == "kogito")
}

func Test_addDefaultMeta_whenAlreadyHasAnnotation(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{
		Annotations: map[string]string{
			"test": "test",
		},
	}
	addDefaultMeta(objectMeta, kogitoApp)
	assert.True(t, objectMeta.Annotations["test"] == "test")
	assert.True(t, objectMeta.Annotations["org.kie.kogito/managed-by"] == "Kogito Operator")
}

func Test_addServiceLabels_whenLabelsAreNotProvided(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{}

	kogitoApp = &v1alpha1.KogitoApp{
		Spec: v1alpha1.KogitoAppSpec{
			Name:    "test",
			Service: v1alpha1.KogitoAppServiceObject{},
		},
	}

	addServiceLabels(objectMeta, kogitoApp)
	assert.True(t, objectMeta.Labels[LabelKeyServiceName] == "test")
}

func Test_addServiceLabels_whenAlreadyHasLabels(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{
		Labels: map[string]string{
			"service":  "test123",
			"operator": "kogito",
		},
	}
	addServiceLabels(objectMeta, kogitoApp)
	assert.True(t, objectMeta.Labels[LabelKeyServiceName] == "test")
	assert.True(t, objectMeta.Labels["operator"] == "kogito")
}

func Test_addServiceLabels_whenLabelsAreProvided(t *testing.T) {
	objectMeta := &metav1.ObjectMeta{
		Labels: map[string]string{
			"service":  "test123",
			"operator": "kogito123",
		},
	}

	kogitoApp = &v1alpha1.KogitoApp{
		Spec: v1alpha1.KogitoAppSpec{
			Name: "test",
			Service: v1alpha1.KogitoAppServiceObject{
				Labels: map[string]string{
					"service":  "test456",
					"operator": "kogito456",
				},
			},
		},
	}

	addServiceLabels(objectMeta, kogitoApp)
	assert.True(t, objectMeta.Labels["service"] == "test456")
	assert.True(t, objectMeta.Labels["operator"] == "kogito456")
}
