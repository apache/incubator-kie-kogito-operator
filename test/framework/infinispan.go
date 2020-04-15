package framework

import (
	"fmt"

	v1 "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
)

const (
	InfinispanContainerName = "infinispan"
	ExtraJavaOptions        = "EXTRA_JAVA_OPTIONS"
)

// Configures already running Infinispan CR with the supplied configuration
func ConfigureInfinispan(namespace, name string, config v1.InfinispanContainerSpec) error {
	GetLogger(namespace).Infof("Configure Infinispan instance with extraJvmOpts: '%s', cpu '%s' and memory '%s'", config.ExtraJvmOpts, config.CPU, config.Memory)
	infinispan, err := getInfinispan(namespace, name)
	if err != nil {
		return err
	} else if infinispan == nil {
		return fmt.Errorf("No Infinispan found with name %s in namespace %s", name, namespace)
	}
	infinispan.Spec.Container = config
	return kubernetes.ResourceC(kubeClient).Update(infinispan)
}

// Waits for Infinispan CR to be created
func WaitForInfinispanToBeCreated(namespace, name string, timeoutInMin int) error {
	return WaitForOnOpenshift(namespace, fmt.Sprintf("Infinispan %s created", name), timeoutInMin,
		func() (bool, error) {
			if infinispan, err := getInfinispan(namespace, name); err != nil {
				return false, err
			} else if infinispan == nil {
				return false, nil
			} else {
				return true, nil
			}
		})
}

// Waits for an Infinispan pod with specific labels to be running with expected configuration
func WaitForInfinispanPodToBeRunningWithConfig(namespace string, labels map[string]string, expectedConfig v1.InfinispanContainerSpec, timeoutInMin int) error {
	return WaitForOnOpenshift(namespace, fmt.Sprintf("Infinispan pod with labels '%s' to have resources", labels), timeoutInMin,
		func() (bool, error) {
			pods, err := GetPodsWithLabels(namespace, labels)
			if err != nil {
				return false, err
			} else if pods == nil {
				return false, nil
			}

			for _, pod := range pods.Items {
				// First check that the pod is really running
				for _, condition := range pod.Status.Conditions {
					if condition.Type == corev1.ContainersReady && condition.Status != corev1.ConditionTrue {
						return false, nil
					}
				}
				if !checkContainersResources(pod.Spec.Containers, getExpectedResourceRequirements(expectedConfig.CPU, expectedConfig.Memory)) {
					return false, nil
				}
				if !checkPodContainerHasEnvVariableWithValue(&pod, InfinispanContainerName, ExtraJavaOptions, expectedConfig.ExtraJvmOpts) {
					return false, nil
				}
			}

			return true, nil
		})

}

func getInfinispan(namespace, name string) (*v1.Infinispan, error) {
	infinispan := &v1.Infinispan{}
	if exists, err := kubernetes.ResourceC(kubeClient).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, infinispan); err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("Error while trying to look for Infinispan %s: %v ", name, err)
	} else if errors.IsNotFound(err) || !exists {
		return nil, nil
	}
	return infinispan, nil
}

func getExpectedResourceRequirements(cpu, memory string) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Limits:   corev1.ResourceList{"cpu": resource.MustParse(cpu), "memory":resource.MustParse(memory)},
		Requests: corev1.ResourceList{"cpu": resource.MustParse(cpu), "memory":resource.MustParse(memory)},
	}
}