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

package test

import (
	"strings"
	"testing"

	"github.com/RHsyseng/operator-utils/pkg/validation"
	"github.com/ghodss/yaml"
	packr "github.com/gobuffalo/packr/v2"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestSampleCustomResources(t *testing.T) {
	schema := getSchema(t)
	box := packr.New("deploy/crs", "../../../../deploy/crs")
	for _, file := range box.List() {
		yamlString, err := box.FindString(file)
		assert.NoError(t, err, "Error reading %v CR yaml", file)
		var input map[string]interface{}
		assert.NoError(t, yaml.Unmarshal([]byte(yamlString), &input))
		assert.NoError(t, schema.Validate(input), "File %v does not validate against the CRD schema", file)
	}
}

func TestTrialEnvMinimum(t *testing.T) {
	var inputYaml = `
apiVersion: app.kiegroup.org/v1alpha1
kind: KogitoApp
metadata:
  name: trial
spec:
  build:
    gitSource:
      uri: https://github.com/kiegroup/kogito-examples
`
	var input map[string]interface{}
	assert.NoError(t, yaml.Unmarshal([]byte(inputYaml), &input))

	schema := getSchema(t)
	assert.NoError(t, schema.Validate(input))

	//	deleteNestedMapEntry(input, "spec", "environment")
	//	assert.Error(t, schema.Validate(input))
}

func TestCompleteCRD(t *testing.T) {
	schema := getSchema(t)
	missingEntries := schema.GetMissingEntries(&v1alpha1.KogitoApp{})
	for _, missing := range missingEntries {
		if strings.HasPrefix(missing.Path, "/status") {
			//Not using subresources, so status is not expected to appear in CRD
		} else if strings.Contains(missing.Path, "/envs/valueFrom/") {
			//The valueFrom is not expected to be used and is not fully defined TODO: verify
		} else if strings.HasSuffix(missing.Path, "/from/uid") {
			//The ObjectReference in From is not expected to be used and is not fully defined TODO: verify
		} else if strings.HasSuffix(missing.Path, "/from/apiVersion") {
			//The ObjectReference in From is not expected to be used and is not fully defined TODO: verify
		} else if strings.HasSuffix(missing.Path, "/from/resourceVersion") {
			//The ObjectReference in From is not expected to be used and is not fully defined TODO: verify
		} else if strings.HasSuffix(missing.Path, "/from/fieldPath") {
			//The ObjectReference in From is not expected to be used and is not fully defined TODO: verify
		} else {
			assert.Fail(t, "Discrepancy between CRD and Struct", "Missing or incorrect schema validation at %v, expected type %v", missing.Path, missing.Type)
		}
	}
}

func deleteNestedMapEntry(object map[string]interface{}, keys ...string) {
	for index := 0; index < len(keys)-1; index++ {
		object = object[keys[index]].(map[string]interface{})
	}
	delete(object, keys[len(keys)-1])
}

func getSchema(t *testing.T) validation.Schema {
	box := packr.New("deploy/crds", "../../../../deploy/crds")
	crdFile := "app.kiegroup.org_kogitoapps_crd.yaml"
	assert.True(t, box.Has(crdFile))
	yamlString, err := box.FindString(crdFile)
	assert.NoError(t, err, "Error reading CRD yaml %v", yamlString)
	schema, err := validation.New([]byte(yamlString))
	assert.NoError(t, err)
	return schema
}
