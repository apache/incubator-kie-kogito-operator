package resource

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newStatefulset(instance *v1alpha1.KogitoDataIndex, cm *corev1.ConfigMap, secret corev1.Secret) *appsv1.StatefulSet {
	// create a standard probe
	probe := defaultProbe
	probe.Handler.TCPSocket = &corev1.TCPSocketAction{Port: intstr.FromInt(defaultExposedPort)}
	// environment variables
	removeManagedEnvVars(instance)
	// from cr
	envs := instance.Spec.Env
	// defaults
	envs = util.AppendStringMap(envs, defaultEnvs)
	envs = util.AppendStringMap(envs, fromInfinispanToStringMap(instance.Spec.Infinispan, secret))
	envs = util.AppendStringMap(envs, fromKafkaToStringMap(instance.Spec.Kafka))

	if instance.Spec.Replicas == 0 {
		instance.Spec.Replicas = defaultReplicas
	}
	if len(instance.Spec.Image) == 0 {
		instance.Spec.Image = DefaultImage
	}

	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Spec.Name,
			Namespace: instance.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &instance.Spec.Replicas,
			Selector: &metav1.LabelSelector{},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: defaultProtobufMountName,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cm.Name,
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            instance.Spec.Name,
							Image:           instance.Spec.Image,
							Env:             util.FromMapToEnvVar(envs),
							Resources:       extractResources(instance),
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: defaultExposedPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							LivenessProbe:  probe,
							ReadinessProbe: probe,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      defaultProtobufMountName,
									MountPath: defaultProtobufMountPath,
								},
							},
						},
					},
				},
			},
		},
	}

	meta.SetGroupVersionKind(&statefulset.TypeMeta, meta.KindStatefulSet)
	addDefaultMetadata(&statefulset.ObjectMeta, instance)
	addDefaultMetadata(&statefulset.Spec.Template.ObjectMeta, instance)
	statefulset.Spec.Selector.MatchLabels = statefulset.Labels

	return statefulset
}
