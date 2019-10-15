package compare

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"reflect"
)

type ResourceDelta struct {
	Added   []resource.KubernetesResource
	Updated []resource.KubernetesResource
	Removed []resource.KubernetesResource
}

func (delta *ResourceDelta) HasChanges() bool {
	if len(delta.Added) > 0 {
		return true
	}
	if len(delta.Updated) > 0 {
		return true
	}
	if len(delta.Removed) > 0 {
		return true
	}
	return false
}

type ResourceComparator interface {
	SetDefaultComparator(compFunc func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool)
	GetDefaultComparator() func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool
	SetComparator(resourceType reflect.Type, compFunc func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool)
	GetComparator(resourceType reflect.Type) func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool
	Compare(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool
	CompareArrays(deployed []resource.KubernetesResource, requested []resource.KubernetesResource) ResourceDelta
}

func DefaultComparator() ResourceComparator {
	return &resourceComparator{
		deepEquals,
		defaultMap(),
	}
}

func SimpleComparator() ResourceComparator {
	return &resourceComparator{
		deepEquals,
		make(map[reflect.Type]func(resource.KubernetesResource, resource.KubernetesResource) bool),
	}
}
