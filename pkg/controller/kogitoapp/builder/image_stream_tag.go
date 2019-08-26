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
