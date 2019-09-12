package main

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_DeployCmd_WhenThereAreNoOperator(t *testing.T) {
	cli := fmt.Sprintf("deploy-service example-drools https://github.com/kiegroup/kogito-examples --context-dir drools-quarkus-example --project kogito")
	lines, _, err := executeCli(cli, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kogito"}})
	assert.Error(t, err)
	assert.Contains(t, lines, "Couldn't find the Kogito CRD")
}

func Test_DeployCmd_WhenThereAreAnOperator(t *testing.T) {
	cli := fmt.Sprintf("deploy-service example-drools https://github.com/kiegroup/kogito-examples --context-dir drools-quarkus-example --project kogito")
	lines, _, err := executeCli(cli,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kogito"}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoAppCRDName}},
	)
	assert.NoError(t, err)
	assert.Contains(t, lines, "example-drools")
	assert.Contains(t, lines, "successfully created")
}

func Test_DeployCmd_CustomDeployment(t *testing.T) {
	cli := fmt.Sprintf("deploy-service example-drools https://github.com/kiegroup/kogito-examples -v --context-dir drools-quarkus-example --project kogito --image-s2i=myimage --image-runtime=myimage:0.2 --limits cpu=1 --limits memory=1Gi --requests cpu=1,memory=1Gi")
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
	assert.Equal(t, config.LastKogitoAppCreated.Spec.Build.ImageS2I.ImageStreamName, "myimage")
	assert.Equal(t, config.LastKogitoAppCreated.Spec.Build.ImageRuntime.ImageStreamName, "myimage")
	assert.Equal(t, config.LastKogitoAppCreated.Spec.Build.ImageRuntime.ImageStreamTag, "0.2")
}

func Test_DeployCmd_CustomImage(t *testing.T) {
	cli := fmt.Sprintf("deploy-service example-drools https://github.com/kiegroup/kogito-examples --native=false --context-dir drools-quarkus-example --project kogito --image-s2i=openshift/myimage --image-runtime=openshift/myimage:0.2")
	_, _, err := executeCli(cli,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kogito"}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoAppCRDName}},
	)
	assert.NoError(t, err)

	instance := v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-drools",
			Namespace: "kogito",
		},
	}

	exists, err := kubernetes.ResourceC(kubeCli).Fetch(&instance)
	assert.NoError(t, err)
	assert.True(t, exists)

	assert.Equal(t, "openshift", instance.Spec.Build.ImageS2I.ImageStreamNamespace)
	assert.Equal(t, "myimage", instance.Spec.Build.ImageS2I.ImageStreamName)

	assert.Equal(t, "openshift", instance.Spec.Build.ImageRuntime.ImageStreamNamespace)
	assert.Equal(t, "myimage", instance.Spec.Build.ImageRuntime.ImageStreamName)
	assert.Equal(t, "0.2", instance.Spec.Build.ImageRuntime.ImageStreamTag)
}
