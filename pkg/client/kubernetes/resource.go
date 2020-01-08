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

package kubernetes

import (
	"context"
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/operator"

	runtimecli "sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// ResourceInterface has functions that interacts with any resource object in the Kubernetes cluster
type ResourceInterface interface {
	// Create creates a new Kubernetes object in the cluster.
	// Note that no checks will be performed in the cluster. If you're not sure, use CreateIfNotExists.
	Create(resource meta.ResourceObject) error
	// CreateIfNotExists will fetch for the object resource in the Kubernetes cluster, if not exists, will create it.
	CreateIfNotExists(resource meta.ResourceObject) (exists bool, err error)
	// FetchWithKey fetches and binds a resource from the Kubernetes cluster with the defined key. If not exists, returns false.
	FetchWithKey(key types.NamespacedName, resource meta.ResourceObject) (exists bool, err error)
	// Fetch fetches and binds a resource with given name and namespace from the Kubernetes cluster. If not exists, returns false.
	Fetch(resource meta.ResourceObject) (exists bool, err error)
	// ListWithNamespace fetches and binds a list resource from the Kubernetes cluster with the defined namespace.
	ListWithNamespace(namespace string, list runtime.Object) error
	// ListWithNamespaceAndLabel same as ListWithNamespace, but also limit the query scope by the given labels
	ListWithNamespaceAndLabel(namespace string, list runtime.Object, labels map[string]string) error
	Delete(resource meta.ResourceObject) error
	// UpdateStatus update the given object status
	UpdateStatus(resource meta.ResourceObject) error
	// Update the given object
	Update(resource meta.ResourceObject) error
	// CreateFromYamlContent creates Kubernetes resources from a yaml string content
	CreateFromYamlContent(yamlContent, namespace string, resourceRef meta.ResourceObject, beforeCreate func(object interface{})) error
}

type resource struct {
	client *client.Client
}

func newResource(c *client.Client) *resource {
	if c == nil {
		c = &client.Client{}
	}
	c.ControlCli = client.MustEnsureClient(c)
	return &resource{
		client: c,
	}
}

func (r *resource) UpdateStatus(resource meta.ResourceObject) error {
	log.Debugf("About to update status for object %s on namespace %s", resource.GetName(), resource.GetNamespace())
	if err := r.client.ControlCli.Status().Update(context.TODO(), resource); err != nil {
		return err
	}

	log.Debugf("Object %s status updated. Creation Timestamp: %s", resource.GetName(), resource.GetCreationTimestamp())
	return nil
}

func (r *resource) Update(resource meta.ResourceObject) error {
	log.Debugf("About to update object %s on namespace %s", resource.GetName(), resource.GetNamespace())
	if err := r.client.ControlCli.Update(context.TODO(), resource); err != nil {
		return err
	}

	log.Debugf("Object %s updated. Creation Timestamp: %s", resource.GetName(), resource.GetCreationTimestamp())
	return nil
}

func (r *resource) Fetch(resource meta.ResourceObject) (bool, error) {
	return r.FetchWithKey(types.NamespacedName{Name: resource.GetName(), Namespace: resource.GetNamespace()}, resource)
}

func (r *resource) FetchWithKey(key types.NamespacedName, resource meta.ResourceObject) (bool, error) {
	log.Debugf("About to fetch object '%s' on namespace '%s'", key.Name, key.Namespace)
	err := r.client.ControlCli.Get(context.TODO(), key, resource)
	if err != nil && errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	log.Debugf("Found object (%s) '%s' in the namespace '%s'. Creation time is: %s",
		resource.GetObjectKind().GroupVersionKind().Kind,
		key.Name,
		key.Namespace,
		resource.GetCreationTimestamp())
	return true, nil
}

func (r *resource) Create(resource meta.ResourceObject) error {
	log := log.With("kind", resource.GetObjectKind().GroupVersionKind().Kind, "name", resource.GetName(), "namespace", resource.GetNamespace())
	log.Debug("Creating")
	if err := r.client.ControlCli.Create(context.TODO(), resource); err != nil {
		log.Debug("Failed to create object. ", err)
		return err
	}
	return nil
}

func (r *resource) CreateIfNotExists(resource meta.ResourceObject) (bool, error) {
	log := log.With("kind", resource.GetObjectKind().GroupVersionKind().Kind, "name", resource.GetName(), "namespace", resource.GetNamespace())

	if exists, err := r.Fetch(resource); err == nil && !exists {
		if err := r.Create(resource); err != nil {
			return false, err
		}
		return true, nil
	} else if err != nil {
		log.Debug("Failed to fecth object. ", err)
		return false, err
	}
	log.Debug("Skip creating - object already exists")
	return false, nil
}

func (r *resource) CreateFromYamlContent(yamlFileContent, namespace string, resourceRef meta.ResourceObject, beforeCreate func(object interface{})) error {
	docs := strings.Split(yamlFileContent, "---")
	for _, doc := range docs {
		if err := yaml.NewYAMLOrJSONDecoder(strings.NewReader(doc), len([]byte(doc))).Decode(resourceRef); err != nil {
			return fmt.Errorf("Error while unmarshalling file: %v ", err)
		}

		if namespace != "" {
			resourceRef.SetNamespace(namespace)
		}
		resourceRef.SetResourceVersion("")
		resourceRef.SetLabels(map[string]string{"app": operator.Name})

		log.Debugf("Will create a new resource '%s' with name %s on %s ", resourceRef.GetObjectKind().GroupVersionKind().Kind, resourceRef.GetName(), resourceRef.GetNamespace())
		if beforeCreate != nil {
			beforeCreate(resourceRef)
		}
		if _, err := r.CreateIfNotExists(resourceRef); err != nil {
			return fmt.Errorf("Error creating object %s: %v ", resourceRef.GetName(), err)
		}
	}
	return nil
}

func (r *resource) ListWithNamespace(namespace string, list runtime.Object) error {
	err := r.client.ControlCli.List(context.TODO(), list, runtimecli.InNamespace(namespace))
	if err != nil {
		log.Debug("Failed to list resource. ", err)
		return err
	}
	return nil
}

func (r *resource) ListWithNamespaceAndLabel(namespace string, list runtime.Object, labels map[string]string) error {
	err := r.client.ControlCli.List(context.TODO(), list, runtimecli.InNamespace(namespace), runtimecli.MatchingLabels(labels))
	if err != nil {
		log.Debug("Failed to list resource. ", err)
		return err
	}
	return nil
}

func (r *resource) Delete(resource meta.ResourceObject) error {
	err := r.client.ControlCli.Delete(context.TODO(), resource)
	if err != nil {
		log.Debugf("Failed to delete resource %s", resource.GetName())
		return err
	}
	return nil
}
