package connector

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/framework"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/core/manager"
	"github.com/kiegroup/kogito-cloud-operator/core/operator"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"path"
)

const (
	// Default Proto Buf file path
	defaultProtobufMountPath = operator.KogitoHomeDir + "/data/protobufs"
	// Proto Buf folder env
	protoBufKeyFolder string = "KOGITO_PROTOBUF_FOLDER"
	// Proto Buf watch env
	protoBufKeyWatch string = "KOGITO_PROTOBUF_WATCH"
	// ConfigMapProtoBufEnabledLabelKey label key used by configMaps that are meant to hold protobuf files
	ConfigMapProtoBufEnabledLabelKey = "kogito-protobuf"
)

// ProtoBufHandler ...
type ProtoBufHandler interface {
	MountProtoBufConfigMapsOnDeployment(deployment *appsv1.Deployment) (err error)
	MountProtoBufConfigMapOnDataIndex(kogitoRuntime api.KogitoRuntimeInterface) (err error)
}

type protoBufHandler struct {
	client                   *client.Client
	log                      logger.Logger
	supportingServiceHandler api.KogitoSupportingServiceHandler
}

// NewProtoBufHandler ...
func NewProtoBufHandler(client *client.Client, log logger.Logger, supportingServiceHandler api.KogitoSupportingServiceHandler) ProtoBufHandler {
	return &protoBufHandler{
		client:                   client,
		log:                      log,
		supportingServiceHandler: supportingServiceHandler,
	}
}

// MountProtoBufConfigMapsOnDeployment mounts protobuf configMaps from KogitoRuntime services into the given deployment
func (p *protoBufHandler) MountProtoBufConfigMapsOnDeployment(deployment *appsv1.Deployment) (err error) {
	cms, err := p.getProtoBufConfigMapsForAllRuntimeServices(deployment.Namespace)
	if err != nil || len(cms.Items) == 0 {
		return err
	}
	for _, cm := range cms.Items {
		appendProtoBufVolumeIntoDeployment(deployment, cm)
		appendProtoBufVolumeMountIntoDeployment(deployment, cm)
	}
	updateProtoBufPropInToDeploymentEnv(deployment)
	return nil
}

// MountProtoBufConfigMapOnDataIndex mounts protobuf configMaps from KogitoRuntime services into the given deployment instance of DataIndex
func (p *protoBufHandler) MountProtoBufConfigMapOnDataIndex(kogitoRuntime api.KogitoRuntimeInterface) (err error) {
	supportingServiceManager := manager.NewKogitoSupportingServiceManager(p.client, p.log, p.supportingServiceHandler)
	deployment, err := supportingServiceManager.FetchKogitoSupportingServiceDeployment(kogitoRuntime.GetNamespace(), api.DataIndex)
	if err != nil || deployment == nil {
		return
	}

	cms, err := p.getProtoBufConfigMapsForSpecificRuntimeService(kogitoRuntime)
	if err != nil || len(cms.Items) == 0 {
		return
	}
	for _, cm := range cms.Items {
		appendProtoBufVolumeIntoDeployment(deployment, cm)
		appendProtoBufVolumeMountIntoDeployment(deployment, cm)
	}
	updateProtoBufPropInToDeploymentEnv(deployment)
	return kubernetes.ResourceC(p.client).Update(deployment)
}

func appendProtoBufVolumeIntoDeployment(deployment *appsv1.Deployment, cm corev1.ConfigMap) {
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == cm.Name {
			return
		}
	}

	// append volume if its not exists
	deployment.Spec.Template.Spec.Volumes =
		append(deployment.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: cm.Name,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: &framework.ModeForProtoBufConfigMapVolume,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cm.Name,
					},
				},
			},
		})
}

func appendProtoBufVolumeMountIntoDeployment(deployment *appsv1.Deployment, cm corev1.ConfigMap) {
	for fileName := range cm.Data {
		mountPath := path.Join(defaultProtobufMountPath, cm.Labels["app"], fileName)
		for _, volumeMount := range deployment.Spec.Template.Spec.Containers[0].VolumeMounts {
			if volumeMount.MountPath == mountPath {
				return
			}
		}

		// append volume mount if its not exists
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts =
			append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
				Name:      cm.Name,
				MountPath: mountPath,
				SubPath:   fileName,
			})
	}
}

func updateProtoBufPropInToDeploymentEnv(deployment *appsv1.Deployment) {
	if len(deployment.Spec.Template.Spec.Volumes) > 0 {
		framework.SetEnvVar(protoBufKeyWatch, "true", &deployment.Spec.Template.Spec.Containers[0])
		framework.SetEnvVar(protoBufKeyFolder, defaultProtobufMountPath, &deployment.Spec.Template.Spec.Containers[0])
	} else {
		framework.SetEnvVar(protoBufKeyWatch, "false", &deployment.Spec.Template.Spec.Containers[0])
		framework.SetEnvVar(protoBufKeyFolder, "", &deployment.Spec.Template.Spec.Containers[0])
	}
}

// getProtoBufConfigMapsForAllRuntimeServices will return every configMap labeled as "protobuf=true" in the given namespace
func (p *protoBufHandler) getProtoBufConfigMapsForAllRuntimeServices(namespace string) (*corev1.ConfigMapList, error) {
	cms := &corev1.ConfigMapList{}
	if err := kubernetes.ResourceC(p.client).ListWithNamespaceAndLabel(namespace, cms, map[string]string{ConfigMapProtoBufEnabledLabelKey: "true"}); err != nil {
		return nil, err
	}
	return cms, nil
}

// getProtoBufConfigMapsForAllRuntimeServices will return every configMap labeled as "protobuf=true" in the given namespace
func (p *protoBufHandler) getProtoBufConfigMapsForSpecificRuntimeService(kogitoRuntime api.KogitoRuntimeInterface) (*corev1.ConfigMapList, error) {
	cms := &corev1.ConfigMapList{}
	labelMaps := map[string]string{
		ConfigMapProtoBufEnabledLabelKey: "true",
		framework.LabelAppKey:            kogitoRuntime.GetName(),
	}
	if err := kubernetes.ResourceC(p.client).ListWithNamespaceAndLabel(kogitoRuntime.GetNamespace(), cms, labelMaps); err != nil {
		return nil, err
	}
	return cms, nil
}
