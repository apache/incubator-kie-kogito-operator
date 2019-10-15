package hooks

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	"reflect"
)

type hookFunc = func(existing resource.KubernetesResource, requested resource.KubernetesResource) error

type UpdateHookMap struct {
	DefaultHook hookFunc
	HookMap     map[reflect.Type]hookFunc
}

func DefaultUpdateHooks() *UpdateHookMap {
	hookMap := make(map[reflect.Type]func(existing resource.KubernetesResource, requested resource.KubernetesResource) error)
	hookMap[reflect.TypeOf(corev1.Service{})] = serviceHook
	return &UpdateHookMap{
		DefaultHook: defaultHook,
		HookMap:     hookMap,
	}
}

func (this *UpdateHookMap) Trigger(existing resource.KubernetesResource, requested resource.KubernetesResource) error {
	function := this.HookMap[reflect.ValueOf(existing).Elem().Type()]
	if function == nil {
		function = this.DefaultHook
	}
	return function(existing, requested)
}

func defaultHook(existing resource.KubernetesResource, requested resource.KubernetesResource) error {
	requested.SetResourceVersion(existing.GetResourceVersion())
	requested.GetObjectKind().SetGroupVersionKind(existing.GetObjectKind().GroupVersionKind())
	return nil
}

func serviceHook(existing resource.KubernetesResource, requested resource.KubernetesResource) error {
	existingService := existing.(*corev1.Service)
	requestedService := requested.(*corev1.Service)
	if requestedService.Spec.ClusterIP == "" {
		requestedService.Spec.ClusterIP = existingService.Spec.ClusterIP
	}
	err := defaultHook(existing, requested)
	if err != nil {
		return err
	}
	return nil
}
