package steps

import (
	"github.com/cucumber/godog"
	v1 "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/infinispan"
	"github.com/kiegroup/kogito-cloud-operator/test/framework"
)

var DefaultInfinispanContainerSpec = v1.InfinispanContainerSpec{
	ExtraJvmOpts: "-Xmx2G",
	Memory:       "3Gi",
	CPU:          "1",
}

func registerInfinispanSteps(s *godog.Suite, data *Data) {
	s.Step(`^Infinispan is configured for performance within (\d+) minute\(s\)$`, data.infinispanIsConfiguredForPerformance)
}

func (data *Data) infinispanIsConfiguredForPerformance(timeoutInMin int) error {
	err := framework.WaitForInfinispanToBeCreated(data.Namespace, infinispan.InstanceName, timeoutInMin)
	if err != nil {
		return err
	}
	err = framework.ConfigureInfinispan(data.Namespace, infinispan.InstanceName, DefaultInfinispanContainerSpec)
	if err != nil {
		return err
	}
	return framework.WaitForInfinispanPodToBeRunningWithConfig(data.Namespace, map[string]string{"clusterName": infinispan.InstanceName}, DefaultInfinispanContainerSpec, timeoutInMin)
}
