package kogitoservice

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/framework"
	"github.com/kiegroup/kogito-cloud-operator/core/operator"
	corev1 "k8s.io/api/core/v1"
)

const (
	// AppPropVolumeName is the name of the volume for application.properties
	AppPropVolumeName = "app-prop-config"
	appPropFilePath   = operator.KogitoHomeDir + "/config"
)

// AppPropsVolumeHandler ...
type AppPropsVolumeHandler interface {
	CreateAppPropVolume(service api.KogitoService) corev1.Volume
	CreateAppPropVolumeMount() corev1.VolumeMount
}

type appPropsVolumeHandler struct {
}

// NewAppPropsVolumeHandler ...
func NewAppPropsVolumeHandler() AppPropsVolumeHandler {
	return &appPropsVolumeHandler{}
}

// CreateAppPropVolume creates a volume for application.properties
func (a *appPropsVolumeHandler) CreateAppPropVolume(service api.KogitoService) corev1.Volume {
	return corev1.Volume{
		Name: AppPropVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: getAppPropConfigMapName(service),
				},
				Items: []corev1.KeyToPath{
					{
						Key:  ConfigMapApplicationPropertyKey,
						Path: ConfigMapApplicationPropertyKey,
					},
				},
				DefaultMode: &framework.ModeForPropertyFiles,
			},
		},
	}
}

// CreateAppPropVolumeMount creates a container volume mount for mounting application.properties
func (a *appPropsVolumeHandler) CreateAppPropVolumeMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      AppPropVolumeName,
		MountPath: appPropFilePath,
		ReadOnly:  true,
	}
}
