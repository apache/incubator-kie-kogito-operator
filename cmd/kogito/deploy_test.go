package main

import (
	"fmt"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_DeployCmd_WhenThereAreNoOperator(t *testing.T) {
	cli := fmt.Sprintf("deploy example-drools https://github.com/kiegroup/kogito-examples --context drools-quarkus-example --app kogito")
	lines, _, err := executeCli(cli, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kogito"}})
	assert.Error(t, err)
	assert.Contains(t, lines, "Couldn't find the Kogito Operator")
}

func Test_DeployCmd_WhenThereAreAnOperator(t *testing.T) {
	cli := fmt.Sprintf("deploy example-drools https://github.com/kiegroup/kogito-examples --context drools-quarkus-example --app kogito")
	lines, _, err := executeCli(cli,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kogito"}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoAppCRDName}},
	)
	assert.NoError(t, err)
	assert.Contains(t, lines, "example-drools")
	assert.Contains(t, lines, "successfully created")
}

func Test_DeployCmd_CustomDeployment(t *testing.T) {
	cli := fmt.Sprintf("deploy example-drools https://github.com/kiegroup/kogito-examples -v --context drools-quarkus-example --app kogito --limits cpu=1 --limits memory=1Gi --requests cpu=1,memory=1Gi")
	config.LastKogitoAppCreated = nil
	_, _, err := executeCli(cli,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kogito"}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoAppCRDName}},
	)
	assert.NoError(t, err)
	assert.NotNil(t, config.LastKogitoAppCreated)
	assert.Equal(t, v1alpha1.QuarkusRuntimeType, config.LastKogitoAppCreated.Spec.Runtime)
	assert.Contains(t, config.LastKogitoAppCreated.Spec.Resources.Limits, v1alpha1.ResourceMap{Resource: v1alpha1.ResourceCPU, Value: "1"})
	assert.Contains(t, config.LastKogitoAppCreated.Spec.Resources.Requests, v1alpha1.ResourceMap{Resource: v1alpha1.ResourceMemory, Value: "1Gi"})
}
