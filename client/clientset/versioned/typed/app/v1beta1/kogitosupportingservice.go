// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
// Code generated by client-gen. DO NOT EDIT.

package v1beta1

import (
	"context"
	"time"

	v1beta1 "github.com/apache/incubator-kie-kogito-operator/apis/app/v1beta1"
	scheme "github.com/apache/incubator-kie-kogito-operator/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// KogitoSupportingServicesGetter has a method to return a KogitoSupportingServiceInterface.
// A group's client should implement this interface.
type KogitoSupportingServicesGetter interface {
	KogitoSupportingServices(namespace string) KogitoSupportingServiceInterface
}

// KogitoSupportingServiceInterface has methods to work with KogitoSupportingService resources.
type KogitoSupportingServiceInterface interface {
	Create(ctx context.Context, kogitoSupportingService *v1beta1.KogitoSupportingService, opts v1.CreateOptions) (*v1beta1.KogitoSupportingService, error)
	Update(ctx context.Context, kogitoSupportingService *v1beta1.KogitoSupportingService, opts v1.UpdateOptions) (*v1beta1.KogitoSupportingService, error)
	UpdateStatus(ctx context.Context, kogitoSupportingService *v1beta1.KogitoSupportingService, opts v1.UpdateOptions) (*v1beta1.KogitoSupportingService, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1beta1.KogitoSupportingService, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1beta1.KogitoSupportingServiceList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.KogitoSupportingService, err error)
	KogitoSupportingServiceExpansion
}

// kogitoSupportingServices implements KogitoSupportingServiceInterface
type kogitoSupportingServices struct {
	client rest.Interface
	ns     string
}

// newKogitoSupportingServices returns a KogitoSupportingServices
func newKogitoSupportingServices(c *AppV1beta1Client, namespace string) *kogitoSupportingServices {
	return &kogitoSupportingServices{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the kogitoSupportingService, and returns the corresponding kogitoSupportingService object, and an error if there is any.
func (c *kogitoSupportingServices) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.KogitoSupportingService, err error) {
	result = &v1beta1.KogitoSupportingService{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kogitosupportingservices").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of KogitoSupportingServices that match those selectors.
func (c *kogitoSupportingServices) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.KogitoSupportingServiceList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta1.KogitoSupportingServiceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kogitosupportingservices").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested kogitoSupportingServices.
func (c *kogitoSupportingServices) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("kogitosupportingservices").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a kogitoSupportingService and creates it.  Returns the server's representation of the kogitoSupportingService, and an error, if there is any.
func (c *kogitoSupportingServices) Create(ctx context.Context, kogitoSupportingService *v1beta1.KogitoSupportingService, opts v1.CreateOptions) (result *v1beta1.KogitoSupportingService, err error) {
	result = &v1beta1.KogitoSupportingService{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("kogitosupportingservices").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kogitoSupportingService).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a kogitoSupportingService and updates it. Returns the server's representation of the kogitoSupportingService, and an error, if there is any.
func (c *kogitoSupportingServices) Update(ctx context.Context, kogitoSupportingService *v1beta1.KogitoSupportingService, opts v1.UpdateOptions) (result *v1beta1.KogitoSupportingService, err error) {
	result = &v1beta1.KogitoSupportingService{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("kogitosupportingservices").
		Name(kogitoSupportingService.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kogitoSupportingService).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *kogitoSupportingServices) UpdateStatus(ctx context.Context, kogitoSupportingService *v1beta1.KogitoSupportingService, opts v1.UpdateOptions) (result *v1beta1.KogitoSupportingService, err error) {
	result = &v1beta1.KogitoSupportingService{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("kogitosupportingservices").
		Name(kogitoSupportingService.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kogitoSupportingService).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the kogitoSupportingService and deletes it. Returns an error if one occurs.
func (c *kogitoSupportingServices) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kogitosupportingservices").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *kogitoSupportingServices) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kogitosupportingservices").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched kogitoSupportingService.
func (c *kogitoSupportingServices) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.KogitoSupportingService, err error) {
	result = &v1beta1.KogitoSupportingService{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("kogitosupportingservices").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
