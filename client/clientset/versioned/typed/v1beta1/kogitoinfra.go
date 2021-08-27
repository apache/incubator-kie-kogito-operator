// Copyright 2021 Red Hat, Inc. and/or its affiliates
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
// Code generated by client-gen. DO NOT EDIT.

package v1beta1

import (
	"context"
	"time"

	v1beta1 "github.com/kiegroup/kogito-operator/api/v1beta1"
	scheme "github.com/kiegroup/kogito-operator/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// KogitoInfrasGetter has a method to return a KogitoInfraInterface.
// A group's client should implement this interface.
type KogitoInfrasGetter interface {
	KogitoInfras(namespace string) KogitoInfraInterface
}

// KogitoInfraInterface has methods to work with KogitoInfra resources.
type KogitoInfraInterface interface {
	Create(ctx context.Context, kogitoInfra *v1beta1.KogitoInfra, opts v1.CreateOptions) (*v1beta1.KogitoInfra, error)
	Update(ctx context.Context, kogitoInfra *v1beta1.KogitoInfra, opts v1.UpdateOptions) (*v1beta1.KogitoInfra, error)
	UpdateStatus(ctx context.Context, kogitoInfra *v1beta1.KogitoInfra, opts v1.UpdateOptions) (*v1beta1.KogitoInfra, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1beta1.KogitoInfra, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1beta1.KogitoInfraList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.KogitoInfra, err error)
	KogitoInfraExpansion
}

// kogitoInfras implements KogitoInfraInterface
type kogitoInfras struct {
	client rest.Interface
	ns     string
}

// newKogitoInfras returns a KogitoInfras
func newKogitoInfras(c *V1beta1Client, namespace string) *kogitoInfras {
	return &kogitoInfras{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the kogitoInfra, and returns the corresponding kogitoInfra object, and an error if there is any.
func (c *kogitoInfras) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.KogitoInfra, err error) {
	result = &v1beta1.KogitoInfra{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kogitoinfras").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of KogitoInfras that match those selectors.
func (c *kogitoInfras) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.KogitoInfraList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta1.KogitoInfraList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("kogitoinfras").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested kogitoInfras.
func (c *kogitoInfras) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("kogitoinfras").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a kogitoInfra and creates it.  Returns the server's representation of the kogitoInfra, and an error, if there is any.
func (c *kogitoInfras) Create(ctx context.Context, kogitoInfra *v1beta1.KogitoInfra, opts v1.CreateOptions) (result *v1beta1.KogitoInfra, err error) {
	result = &v1beta1.KogitoInfra{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("kogitoinfras").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kogitoInfra).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a kogitoInfra and updates it. Returns the server's representation of the kogitoInfra, and an error, if there is any.
func (c *kogitoInfras) Update(ctx context.Context, kogitoInfra *v1beta1.KogitoInfra, opts v1.UpdateOptions) (result *v1beta1.KogitoInfra, err error) {
	result = &v1beta1.KogitoInfra{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("kogitoinfras").
		Name(kogitoInfra.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kogitoInfra).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *kogitoInfras) UpdateStatus(ctx context.Context, kogitoInfra *v1beta1.KogitoInfra, opts v1.UpdateOptions) (result *v1beta1.KogitoInfra, err error) {
	result = &v1beta1.KogitoInfra{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("kogitoinfras").
		Name(kogitoInfra.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(kogitoInfra).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the kogitoInfra and deletes it. Returns an error if one occurs.
func (c *kogitoInfras) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kogitoinfras").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *kogitoInfras) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("kogitoinfras").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched kogitoInfra.
func (c *kogitoInfras) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.KogitoInfra, err error) {
	result = &v1beta1.KogitoInfra{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("kogitoinfras").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
