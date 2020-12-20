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
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/operator"

	"k8s.io/apimachinery/pkg/util/yaml"
)

// ResourceInterface has functions that interacts with any resource object in the Kubernetes cluster
type ResourceInterface interface {
	ResourceReader
	ResourceWriter
	// CreateIfNotExists will fetch for the object resource in the Kubernetes cluster, if not exists, will create it.
	CreateIfNotExists(resource meta.ResourceObject) (err error)
	// CreateIfNotExistsForOwner sets the controller owner to the given resource and creates if it not exists.
	// If the given resource exists, won't update the object with the given owner.
	CreateIfNotExistsForOwner(resource meta.ResourceObject, owner metav1.Object, scheme *runtime.Scheme) (err error)
	// CreateForOwner sets the controller owner to the given resource and creates the resource.
	CreateForOwner(resource meta.ResourceObject, owner metav1.Object, scheme *runtime.Scheme) error
	// CreateFromYamlContent creates Kubernetes resources from a yaml string content
	CreateFromYamlContent(yamlContent, namespace string, resourceRef meta.ResourceObject, beforeCreate func(object interface{})) error
}

type resource struct {
	ResourceReader
	ResourceWriter
}

func newResource(c *client.Client) *resource {
	if c == nil {
		c = &client.Client{}
	}
	c.ControlCli = client.MustEnsureClient(c)
	return &resource{
		ResourceReader: ResourceReaderC(c),
		ResourceWriter: ResourceWriterC(c),
	}
}

func (r *resource) CreateIfNotExists(resource meta.ResourceObject) error {
	log.Info("Create resource if not exists", "kind", resource.GetObjectKind().GroupVersionKind().Kind, "name", resource.GetName(), "namespace", resource.GetNamespace())

	if exists, err := r.ResourceReader.Fetch(resource); err == nil && !exists {
		if err := r.ResourceWriter.Create(resource); err != nil {
			return err
		}
		return nil
	} else if err != nil {
		log.Error(err, "Failed to fetch object. ")
		return err
	}
	log.Info("Skip creating - object already exists")
	return nil
}

func (r *resource) CreateIfNotExistsForOwner(resource meta.ResourceObject, owner metav1.Object, scheme *runtime.Scheme) error {
	err := controllerutil.SetControllerReference(owner, resource, scheme)
	if err != nil {
		return err
	}
	return r.CreateIfNotExists(resource)
}

func (r *resource) CreateForOwner(resource meta.ResourceObject, owner metav1.Object, scheme *runtime.Scheme) error {
	err := controllerutil.SetControllerReference(owner, resource, scheme)
	if err != nil {
		return err
	}
	if err := r.ResourceWriter.Create(resource); err != nil {
		return err
	}
	return nil
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

		log.Debug("Will create a new resource", "kind", resourceRef.GetObjectKind().GroupVersionKind().Kind, "name", resourceRef.GetName(), "namespace", resourceRef.GetNamespace())
		if beforeCreate != nil {
			beforeCreate(resourceRef)
		}
		if err := r.CreateIfNotExists(resourceRef); err != nil {
			return fmt.Errorf("Error creating object %s: %v ", resourceRef.GetName(), err)
		}
	}
	return nil
}
