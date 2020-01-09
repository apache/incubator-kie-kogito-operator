// Copyright 2019 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package framework

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
)

// TODO: remove this since now we're relying on Operator Utils way of doing things

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
