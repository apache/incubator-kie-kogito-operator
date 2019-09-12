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

package builder

import (
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	imgv1 "github.com/openshift/api/image/v1"
)

// NewImageStreamTag creates a new ImageStreamTag on the OpenShift cluster using the tag reference name.
// tagRefName refers to a full tag name like kogito-app:latest. If no tag is passed (e.g. kogito-app), "latest" will be used for the tag
func NewImageStreamTag(kogitoApp *v1alpha1.KogitoApp, tagRefName string) *imgv1.ImageStream {
	result := strings.Split(tagRefName, ":")
	if len(result) == 1 {
		result = append(result, openshift.ImageTagLatest)
	}

	is := &imgv1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:      result[0],
			Namespace: kogitoApp.Namespace,
		},
		Spec: imgv1.ImageStreamSpec{
			LookupPolicy: imgv1.ImageLookupPolicy{
				Local: true,
			},
			Tags: []imgv1.TagReference{
				{
					Name: result[1],
					ReferencePolicy: imgv1.TagReferencePolicy{
						Type: imgv1.LocalTagReferencePolicy,
					},
				},
			},
		},
	}

	addDefaultMeta(&is.ObjectMeta, kogitoApp)
	meta.SetGroupVersionKind(&is.TypeMeta, meta.KindImageStream)

	return is
}
