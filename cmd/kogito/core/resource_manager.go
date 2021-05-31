// Copyright 2021 Red Hat, Inc. and/or its affiliates
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

package core

import (
	kogitocli "github.com/kiegroup/kogito-operator/core/client"
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"sync"
	"time"
)

const (
	cancelUpdateTimeout = 30 * time.Second
	poolWaitTimeout     = 500 * time.Millisecond
)

// ResourceManager is the API entry point for Kubernetes Resource management for CLI
type ResourceManager interface {
	CreateOrUpdate(resource kubernetes.ResourceObject) error
}

// NewResourceManager creates a new reference of a ResourceManager
func NewResourceManager(client *kogitocli.Client) ResourceManager {
	return &resourceManager{Client: client}
}

type resourceManager struct {
	*kogitocli.Client
}

// CreateOrUpdate creates the Kubernetes Resource if it does not exist, updates otherwise
func (r *resourceManager) CreateOrUpdate(resource kubernetes.ResourceObject) error {
	fetchedResource := resource.DeepCopyObject()
	exists, err := kubernetes.ResourceC(r.Client).Fetch(fetchedResource.(kubernetes.ResourceObject))
	if err != nil {
		return err
	}
	if exists {
		return r.update(resource, fetchedResource.(kubernetes.ResourceObject))
	}
	return kubernetes.ResourceC(r.Client).Create(resource)
}

func (r *resourceManager) update(newResource kubernetes.ResourceObject, oldResource kubernetes.ResourceObject) error {
	var updateError error
	var wg sync.WaitGroup
	wg.Add(1)
	go func(res kubernetes.ResourceObject) {
		defer wg.Done()
		// handle race conditions
		err := wait.Poll(poolWaitTimeout, cancelUpdateTimeout, func() (bool, error) {
			res.SetResourceVersion(oldResource.GetResourceVersion())
			err := kubernetes.ResourceC(r.Client).Update(res)
			if err == nil {
				return true, nil
			} else if errors.IsConflict(err) {
				_, err := kubernetes.ResourceC(r.Client).Fetch(oldResource)
				if err != nil {
					return false, err
				}
				return false, nil
			}
			return true, err
		})
		if err != nil {
			updateError = err
			return
		}
	}(newResource)
	wg.Wait()
	return updateError
}
