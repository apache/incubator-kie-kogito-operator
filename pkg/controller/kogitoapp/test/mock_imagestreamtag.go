package test

import (
	"fmt"

	imagev1 "github.com/openshift/api/image/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MockImageStreamTag struct {
	Tags map[string]*imagev1.ImageStreamTag
}

func (mock *MockImageStreamTag) Create(tag *imagev1.ImageStreamTag) (*imagev1.ImageStreamTag, error) {
	if mock.Tags == nil {
		mock.Tags = make(map[string]*imagev1.ImageStreamTag)
	}
	name := fmt.Sprintf("%s/%s", tag.ObjectMeta.Namespace, tag.ObjectMeta.Name)
	mock.Tags[name] = tag
	return tag, nil
}

func (mock *MockImageStreamTag) Update(tag *imagev1.ImageStreamTag) (*imagev1.ImageStreamTag, error) {
	if mock.Tags == nil {
		mock.Tags = make(map[string]*imagev1.ImageStreamTag)
	}
	name := fmt.Sprintf("%s/%s", tag.ObjectMeta.Namespace, tag.ObjectMeta.Name)
	old := mock.Tags[name]
	mock.Tags[name] = tag
	return old, nil
}

func (mock *MockImageStreamTag) Delete(name string, options *meta_v1.DeleteOptions) error {
	if mock.Tags == nil {
		return nil
	}
	delete(mock.Tags, name)
	return nil
}

func (mock *MockImageStreamTag) Get(name string, options meta_v1.GetOptions) (*imagev1.ImageStreamTag, error) {
	if mock.Tags == nil {
		return nil, nil
	}
	return mock.Tags[name], nil
}
