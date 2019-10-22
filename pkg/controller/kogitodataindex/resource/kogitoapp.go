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

package resource

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	"github.com/openshift/api/image/docker10"
	"k8s.io/apimachinery/pkg/types"
)

// getAllProtoFilesFromKogitoApps returns every single protofile attached in every kogitoapp deployed within the namespace
func getAllProtoFilesFromKogitoApps(client *client.Client, namespace string) (map[string]string, error) {
	kogitoApps := &v1alpha1.KogitoAppList{}
	if err := kubernetes.ResourceC(client).ListWithNamespace(namespace, kogitoApps); err != nil {
		return nil, err
	}
	files := map[string]string{}
	for _, item := range kogitoApps.Items {
		var dockerImage *docker10.DockerImage
		var err error
		if dockerImage, err = openshift.ImageStreamC(client).FetchDockerImage(types.NamespacedName{
			Namespace: namespace,
			Name:      item.Name,
		}); err != nil {
			return nil, err
		}
		if dockerImage != nil {
			files = util.AppendStringMap(files, resource.ExtractProtoBufFilesFromDockerImage(item.Name, dockerImage))
		}
	}
	return files, nil
}
