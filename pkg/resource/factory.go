package resource

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
)

// FactoryContext is the base structure needed to create the resources objects in the cluster
type FactoryContext struct {
	Client *client.Client
	// PostCreate is a function that will be called after building the object in the cluster
	PostCreate func(object meta.ResourceObject) error
	// PreCreate is a function called before the object persistence in the cluster
	PreCreate func(object meta.ResourceObject) error
}

// Factory will provide resources to chain resources builders
type Factory struct {
	Context *FactoryContext
	Error   error
}

// CallPostCreate will call a post creation hook
func (f *Factory) CallPostCreate(isNew bool, object meta.ResourceObject) *Factory {
	if isNew && f.Context.PostCreate != nil {
		if f.Error == nil {
			f.Error = f.Context.PostCreate(object)
		}
	}
	return f
}

// CallPreCreate will call a pre creation hook
func (f *Factory) CallPreCreate(object meta.ResourceObject) error {
	if f.Error == nil && f.Context.PreCreate != nil {
		return f.Context.PreCreate(object)
	}
	return nil
}
