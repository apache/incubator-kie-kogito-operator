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

func Test_DeployDataIndexCmd_WhenThereAreNoOperator(t *testing.T) {
	cli := fmt.Sprintf("deploy-data-index --project kogito --infinispan-url myservice:11222 --kafka-url my-cluster:9092")
	lines, _, err := executeCli(cli, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kogito"}})
	assert.Error(t, err)
	assert.Contains(t, lines, "Couldn't find the Kogito CRD")
}

func Test_DeployDataIndexCmd_RequiredFlags(t *testing.T) {
	cli := fmt.Sprintf("deploy-data-index --project kogito")
	lines, _, err := executeCli(cli, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kogito"}})
	assert.Error(t, err)
	assert.Contains(t, lines, "required flag(s) \"infinispan-url\", \"kafka-url\" not set")
}

func Test_DeployDataIndexCmd_SuccessfullDeploy(t *testing.T) {
	cli := fmt.Sprintf("deploy-data-index --project kogito --infinispan-url myservice:11222 --kafka-url my-cluster:9092")
	lines, _, err := executeCli(cli,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kogito"}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoDataIndexCRDName}})
	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Data Index Service successfully deployed")
}

func Test_DeployDataIndexCmd_SuccessfullDeployWithInfinispanCredentials(t *testing.T) {
	cli := fmt.Sprintf("deploy-data-index --project kogito --infinispan-url myservice:11222 --kafka-url my-cluster:9092 --infinispan-user user --infinispan-password password")
	lines, _, err := executeCli(cli,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kogito"}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoDataIndexCRDName}})
	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Data Index Service successfully deployed")
}

func Test_DeployDataIndexCmd_SuccessfullDeployWithInfinispanCredentialsAndSecret(t *testing.T) {
	cli := fmt.Sprintf("deploy-data-index --project kogito --infinispan-url myservice:11222 --kafka-url my-cluster:9092 --infinispan-user user --infinispan-password password")
	lines, _, err := executeCli(cli,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kogito"}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: v1alpha1.KogitoDataIndexCRDName}},
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: defaultInfinispanSecretName, Namespace: "kogito"}})
	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Data Index Service successfully deployed")

	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{
		Name:      defaultInfinispanSecretName,
		Namespace: "kogito",
	}}
	exists, err := kubernetes.ResourceC(kubeCli).Fetch(secret)
	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.True(t, exists)
	assert.Contains(t, secret.StringData, defaulInfinispanUsernameKey, defaultInfinispanPasswordKey)
}
