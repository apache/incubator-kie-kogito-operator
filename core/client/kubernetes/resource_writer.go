// Copyright 2020 Red Hat, Inc. and/or its affiliates
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

package kubernetes

import (
	"context"
	"github.com/RHsyseng/operator-utils/pkg/resource/write"
	"github.com/kiegroup/kogito-operator/core/client"
	runtime "sigs.k8s.io/controller-runtime/pkg/client"
)

// ResourceWriter interface to write kubernetes object
type ResourceWriter interface {
	// Create creates a new Kubernetes object in the cluster.
	// Note that no checks will be performed in the cluster. If you're not sure, use CreateIfNotExists.
	Create(resource runtime.Object) error
	// Delete delete the given object
	Delete(resource runtime.Object) error
	// Update the given object
	Update(resource runtime.Object) error
	// UpdateStatus update the given object status
	UpdateStatus(resource runtime.Object) error
	// CreateResources create provided objects
	CreateResources(resources []runtime.Object) (bool, error)
	// UpdateResources update provided objects
	UpdateResources(existing []runtime.Object, resources []runtime.Object) (bool, error)
	// DeleteResources delete provided objects
	DeleteResources(resources []runtime.Object) (bool, error)
}

// ResourceWriterC provide ResourceWrite reference
func ResourceWriterC(cli *client.Client) ResourceWriter {
	return &resourceWriter{
		client: cli,
	}
}

type resourceWriter struct {
	client *client.Client
}

func (r *resourceWriter) Create(resource runtime.Object) error {
	log.Debug("Creating resource", "kind", resource.GetObjectKind().GroupVersionKind().Kind, "name", resource.GetName(), "namespace", resource.GetNamespace())
	if err := r.client.ControlCli.Create(context.TODO(), resource); err != nil {
		log.Error(err, "Failed to create object. ")
		return err
	}
	return nil
}

func (r *resourceWriter) Update(resource runtime.Object) error {
	log.Debug("About to update resource", "name", resource.GetName(), "namespace", resource.GetNamespace())
	if err := r.client.ControlCli.Update(context.TODO(), resource); err != nil {
		return err
	}
	log.Debug("Resource updated.", "name", resource.GetName(), "Creation Timestamp", resource.GetCreationTimestamp(), "Resource", resource)
	return nil
}

func (r *resourceWriter) Delete(resource runtime.Object) error {
	if err := r.client.ControlCli.Delete(context.TODO(), resource); err != nil {
		log.Error(err, "Failed to delete resource.", "name", resource.GetName())
		return err
	}
	return nil
}

func (r *resourceWriter) UpdateStatus(resource runtime.Object) error {
	log.Debug("About to update status for object", "name", resource.GetName(), "namespace", resource.GetNamespace())
	if err := r.client.ControlCli.Status().Update(context.TODO(), resource); err != nil {
		return err
	}

	log.Debug("Object status updated.", "name", resource.GetName(), "Creation Timestamp", resource.GetCreationTimestamp())
	return nil
}

func (r *resourceWriter) CreateResources(resources []runtime.Object) (bool, error) {
	writer := write.New(r.client.ControlCli)
	return writer.AddResources(resources)
}

func (r *resourceWriter) UpdateResources(existing []runtime.Object, resources []runtime.Object) (bool, error) {
	writer := write.New(r.client.ControlCli)
	return writer.UpdateResources(existing, resources)
}

func (r *resourceWriter) DeleteResources(resources []runtime.Object) (bool, error) {
	writer := write.New(r.client.ControlCli)
	return writer.RemoveResources(resources)
}
