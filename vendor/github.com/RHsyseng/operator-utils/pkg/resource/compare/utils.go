package compare

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"reflect"
)

type mapBuilder struct {
	resourceMap map[reflect.Type][]resource.KubernetesResource
}

func NewMapBuilder() *mapBuilder {
	this := &mapBuilder{resourceMap: make(map[reflect.Type][]resource.KubernetesResource)}
	return this
}

func (this *mapBuilder) ResourceMap() map[reflect.Type][]resource.KubernetesResource {
	return this.resourceMap
}

func (this *mapBuilder) Add(resources ...resource.KubernetesResource) *mapBuilder {
	for index := range resources {
		if resources[index] == nil || reflect.ValueOf(resources[index]).IsNil() {
			continue
		}
		resourceType := reflect.ValueOf(resources[index]).Elem().Type()
		this.resourceMap[resourceType] = append(this.resourceMap[resourceType], resources[index])
	}
	return this
}
