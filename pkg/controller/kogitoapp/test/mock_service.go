package test

import (
	"context"

	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/logs"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
)

var log = logs.GetLogger("kieapp.test")

type MockPlatformService struct {
	CreateFunc          func(ctx context.Context, obj runtime.Object) error
	GetFunc             func(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error
	ListFunc            func(ctx context.Context, opts *clientv1.ListOptions, list runtime.Object) error
	UpdateFunc          func(ctx context.Context, obj runtime.Object) error
	UpdateStatusFunc    func(ctx context.Context, obj runtime.Object) error
	GetCachedFunc       func(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error
	ImageStreamTagsFunc func(namespace string) imagev1.ImageStreamTagInterface
	GetSchemeFunc       func() *runtime.Scheme
}

func MockService() *MockPlatformService {
	mockImageStreamTag := &MockImageStreamTag{}
	return &MockPlatformService{
		CreateFunc: func(ctx context.Context, obj runtime.Object) error {
			log.Debugf("Mock service will do no-op in lieu of creating %v", obj)
			return nil
		},
		GetFunc: func(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
			log.Debugf("Will return nil to request to get key %v and object %v")
			return nil
		},
		ListFunc: func(ctx context.Context, opts *clientv1.ListOptions, list runtime.Object) error {
			return nil
		},
		UpdateFunc: func(ctx context.Context, obj runtime.Object) error {
			log.Debugf("Mock service will do no-op in lieu of updating %v", obj)
			return nil
		},
		UpdateStatusFunc: func(ctx context.Context, obj runtime.Object) error {
			log.Debugf("Mock service will do no-op in lieu of updating status %v", obj)
			return nil
		},
		GetCachedFunc: func(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
			return nil
		},
		ImageStreamTagsFunc: func(namespace string) imagev1.ImageStreamTagInterface {
			return mockImageStreamTag
		},
		GetSchemeFunc: func() *runtime.Scheme {
			return nil
		},
	}
}

func (service *MockPlatformService) Create(ctx context.Context, obj runtime.Object) error {
	return service.CreateFunc(ctx, obj)
}

func (service *MockPlatformService) Get(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
	return service.GetFunc(ctx, key, obj)
}

func (service *MockPlatformService) List(ctx context.Context, opts *clientv1.ListOptions, list runtime.Object) error {
	return service.ListFunc(ctx, opts, list)
}

func (service *MockPlatformService) Update(ctx context.Context, obj runtime.Object) error {
	return service.UpdateFunc(ctx, obj)
}

func (service *MockPlatformService) GetCached(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
	return service.GetCachedFunc(ctx, key, obj)
}

func (service *MockPlatformService) ImageStreamTags(namespace string) imagev1.ImageStreamTagInterface {
	return service.ImageStreamTagsFunc(namespace)
}

func (service *MockPlatformService) GetScheme() *runtime.Scheme {
	return service.GetSchemeFunc()
}

func (service *MockPlatformService) IsMockService() bool {
	return true
}
