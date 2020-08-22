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
	CreateIfNotExists(resource meta.ResourceObject) (exists bool, err error)
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

func (r *resource) CreateIfNotExists(resource meta.ResourceObject) (bool, error) {
	log := log.With("kind", resource.GetObjectKind().GroupVersionKind().Kind, "name", resource.GetName(), "namespace", resource.GetNamespace())

	if exists, err := r.ResourceReader.Fetch(resource); err == nil && !exists {
		if err := r.ResourceWriter.Create(resource); err != nil {
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
