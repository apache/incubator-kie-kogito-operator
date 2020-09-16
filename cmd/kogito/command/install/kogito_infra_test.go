package install

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_InstallInfraServiceCmd_DefaultConfiguration(t *testing.T) {
	name := "kogito-infinispan-infra"
	ns := t.Name()
	cli := fmt.Sprintf("install infra %s --project %s --apiVersion app.infinispan.org/v1 --kind Infinispan", name, ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&v1alpha1.KogitoInfra{ObjectMeta: metav1.ObjectMeta{Name: "kogito-infra", Namespace: ns}})
	lines, _, err := test.ExecuteCli()

	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Infra Service successfully installed")

	// This should be created, given the command above
	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(ctx.Client).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.NotNil(t, kogitoInfra)
	assert.Equal(t, name, kogitoInfra.Name)
	assert.Equal(t, ns, kogitoInfra.Namespace)
	assert.Equal(t, "app.infinispan.org/v1", kogitoInfra.Spec.Resource.APIVersion)
	assert.Equal(t, "Infinispan", kogitoInfra.Spec.Resource.Kind)
}
