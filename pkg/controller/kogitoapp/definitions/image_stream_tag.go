package definitions

import (
	"fmt"
	"strings"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	imgv1 "github.com/openshift/api/image/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	//ImageTagLatest is the default name for the latest image tag
	ImageTagLatest = "latest"
)

// NewImageStreamTag creates a new ImageStreamTag on the OpenShift cluster using the tag reference name.
// tagRefName refers to a full tag name like kogito-app:latest. If no tag is passed (e.g. kogito-app), "latest" will be used for the tag
func NewImageStreamTag(kogitoApp *v1alpha1.KogitoApp, tagRefName string) *imgv1.ImageStreamTag {
	result := strings.Split(tagRefName, ":")
	if len(result) == 1 {
		result = append(result, ImageTagLatest)
	}

	is := &imgv1.ImageStreamTag{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s:%s", result[0], result[1]),
			Namespace: kogitoApp.Namespace,
		},
		Tag: &imgv1.TagReference{
			Name: result[1],
			ReferencePolicy: imgv1.TagReferencePolicy{
				Type: imgv1.LocalTagReferencePolicy,
			},
		},
	}

	addDefaultMeta(&is.ObjectMeta, kogitoApp)
	SetGroupVersionKind(&is.TypeMeta, KindImageStreamTag)

	return is
}
