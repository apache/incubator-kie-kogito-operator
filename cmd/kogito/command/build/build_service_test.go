package build

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

// kogitoBuildCRDName is the name of the Kogito Build CRD in the cluster
const kogitoBuildCRDName = "kogitobuild.app.kiegroup.org"

func Test_BuildServiceCmd_DefaultConfigurations(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("build-service example-quarkus https://github.com/kiegroup/kogito-examples/ --context-dir=drools-quarkus-example --project %s", ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: kogitoBuildCRDName}})

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Build Service successfully installed")

	// This should be created, given the command above
	kogitoBuild := &v1alpha1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-quarkus",
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(ctx.Client).Fetch(kogitoBuild)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, v1alpha1.RemoteSourceBuildType, kogitoBuild.Spec.Type)
	assert.Equal(t, false, kogitoBuild.Spec.DisableIncremental)
	assert.Equal(t, v1alpha1.QuarkusRuntimeType, kogitoBuild.Spec.Runtime)
	assert.Equal(t, false, kogitoBuild.Spec.EnableMavenDownloadOutput)
}

func Test_BuildServiceCmd_CustomConfigurations(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("build example-quarkus https://github.com/kiegroup/kogito-examples/ --context-dir=drools-quarkus-example --runtime=springboot --build-image=quay.io/vajain/kogito-springboot-ubi8-s2i:2.0 --runtime-image=quay.io/vajain/kogito-springboot-ubi8:1.0 --maven-mirror-url=http://172.18.0.1:8080/repository/local/ --maven-output=true --project %s", ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&apiextensionsv1beta1.CustomResourceDefinition{ObjectMeta: metav1.ObjectMeta{Name: kogitoBuildCRDName}})

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Build Service successfully installed")

	// This should be created, given the command above
	kogitoBuild := &v1alpha1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-quarkus",
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(ctx.Client).Fetch(kogitoBuild)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, v1alpha1.RemoteSourceBuildType, kogitoBuild.Spec.Type)
	assert.Equal(t, false, kogitoBuild.Spec.DisableIncremental)
	assert.Equal(t, v1alpha1.SpringbootRuntimeType, kogitoBuild.Spec.Runtime)
	assert.Equal(t, "quay.io/vajain/kogito-springboot-ubi8-s2i:2.0", kogitoBuild.Spec.BuildImage.String())
	assert.Equal(t, "quay.io/vajain/kogito-springboot-ubi8:1.0", kogitoBuild.Spec.RuntimeImage.String())
	assert.Equal(t, "http://172.18.0.1:8080/repository/local/", kogitoBuild.Spec.MavenMirrorURL)
	assert.Equal(t, true, kogitoBuild.Spec.EnableMavenDownloadOutput)
}
