package converter

import (
	"github.com/bmizerany/assert"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"testing"
)

func Test_FromResourceFlagsToResource(t *testing.T) {
	flags := &flag.InfraResourceFlags{
		APIVersion:        "infinispan.org/v1",
		Kind:              "Infinispan",
		ResourceName:      "infinispan-instance-name",
		ResourceNamespace: "infinispan-namespace",
	}

	resource := FromResourceFlagsToResource(flags)
	assert.Equal(t, "infinispan.org/v1", resource.APIVersion)
	assert.Equal(t, "Infinispan", resource.Kind)
	assert.Equal(t, "infinispan-instance-name", resource.Name)
	assert.Equal(t, "infinispan-namespace", resource.Namespace)
}
