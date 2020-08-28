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

package kogitobuild

import (
	"github.com/kiegroup/kogito-cloud-operator/api/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	buildv1 "github.com/openshift/api/build/v1"
	imgv1 "github.com/openshift/api/image/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func getOutputImageStreamNameTag(bc *buildv1.BuildConfig) (name, tag string) {
	imageNameTag := strings.Split(bc.Spec.Output.To.Name, ":")
	name = imageNameTag[0]
	tag = tagLatest
	if len(imageNameTag) > 1 {
		tag = imageNameTag[1]
	}
	return name, tag
}

// newOutputImageStreamForBuilder creates a new output ImageStream for Builder BuildConfigs
func newOutputImageStreamForBuilder(bc *buildv1.BuildConfig) imgv1.ImageStream {
	isName, tag := getOutputImageStreamNameTag(bc)
	return imgv1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:      isName,
			Namespace: bc.Namespace,
			Labels: map[string]string{
				framework.LabelAppKey: bc.Labels[framework.LabelAppKey],
			},
		},
		Spec: imgv1.ImageStreamSpec{
			LookupPolicy: imgv1.ImageLookupPolicy{
				Local: true,
			},
			Tags: []imgv1.TagReference{
				{
					Name: tag,
					ReferencePolicy: imgv1.TagReferencePolicy{
						Type: imgv1.LocalTagReferencePolicy,
					},
				},
			},
		},
	}
}

// newOutputImageStreamForRuntime creates a new image stream for the Runtime
// if one image stream is found in the namespace managed by other resources such as KogitoRuntime or other KogitoBuild, we add ourselves in the owner references
func newOutputImageStreamForRuntime(bc *buildv1.BuildConfig, build *v1alpha1.KogitoBuild, client *client.Client) (*imgv1.ImageStream, error) {
	isName, tag := getOutputImageStreamNameTag(bc)
	sharedImageStream, err := services.GetSharedDeployedImageStream(isName, build.Namespace, client)
	if err != nil {
		return nil, err
	}
	if sharedImageStream != nil {
		return sharedImageStream, nil
	}
	// let's create an ImageStream since we haven't found one in the namespace
	return services.NewImageHandlerForBuiltServices(&v1alpha1.Image{Name: isName, Tag: tag}, build.Namespace, client).GetImageStream(), nil
}
