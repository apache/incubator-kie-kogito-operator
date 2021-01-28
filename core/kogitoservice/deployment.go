package kogitoservice

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/framework"
	"github.com/kiegroup/kogito-cloud-operator/core/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	portName      = "http"
	singleReplica = int32(1)
)

// DeploymentHandler ...
type DeploymentHandler interface {
	CreateRequiredDeployment(service api.KogitoService, resolvedImage string, definition ServiceDefinition) *appsv1.Deployment
	IsDeploymentAvailable(kogitoService api.KogitoService) (bool, error)
}

type deploymentHandler struct {
	client *client.Client
	log    logger.Logger
}

// NewDeploymentHandler ...
func NewDeploymentHandler(client *client.Client, log logger.Logger) DeploymentHandler {
	return &deploymentHandler{
		client: client,
		log:    log,
	}
}

func (d *deploymentHandler) CreateRequiredDeployment(service api.KogitoService, resolvedImage string, definition ServiceDefinition) *appsv1.Deployment {
	if definition.SingleReplica && *service.GetSpec().GetReplicas() > singleReplica {
		service.GetSpec().SetReplicas(singleReplica)
		d.log.Warn("Service can't scale vertically, only one replica is allowed.", "service", service.GetName())
	}
	replicas := service.GetSpec().GetReplicas()
	probes := getProbeForKogitoService(definition, service)
	labels := service.GetSpec().GetDeploymentLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[framework.LabelAppKey] = service.GetName()

	// clone env var slice so that any changes in deployment env var should not reflect in kogitoInstance env var
	// KOGITO-3947: we don't want an empty reference (0 len), since this is nil to k8s. Comparator will go crazy
	var env []corev1.EnvVar
	if len(service.GetSpec().GetEnvs()) > 0 {
		env = make([]corev1.EnvVar, len(service.GetSpec().GetEnvs()))
		copy(env, service.GetSpec().GetEnvs())
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: service.GetName(), Namespace: service.GetNamespace(), Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{framework.LabelAppKey: service.GetName()}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: service.GetName(),
							Ports: []corev1.ContainerPort{
								{
									Name:          portName,
									ContainerPort: int32(framework.DefaultExposedPort),
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env:             env,
							Resources:       service.GetSpec().GetResources(),
							LivenessProbe:   probes.liveness,
							ReadinessProbe:  probes.readiness,
							ImagePullPolicy: corev1.PullAlways,
							Image:           resolvedImage,
						},
					},
				},
			},
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RollingUpdateDeploymentStrategyType},
		},
	}

	return deployment
}

// IsDeploymentAvailable verifies if the Deployment resource from the given KogitoService has replicas available
func (d *deploymentHandler) IsDeploymentAvailable(kogitoService api.KogitoService) (bool, error) {
	// service's deployment hasn't been deployed yet, no need to fetch
	if len(kogitoService.GetStatus().GetDeploymentConditions()) == 0 {
		return false, nil
	}

	coreDeployHandler := infrastructure.NewDeploymentHandler(d.client, d.log)
	deployment, err := coreDeployHandler.FetchDeployment(types.NamespacedName{Name: kogitoService.GetName(), Namespace: kogitoService.GetNamespace()})
	if err != nil {
		return false, err
	} else if deployment == nil {
		return false, nil
	}
	return deployment.Status.AvailableReplicas > 0, nil
}
