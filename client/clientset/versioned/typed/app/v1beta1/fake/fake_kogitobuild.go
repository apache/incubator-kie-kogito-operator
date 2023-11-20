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

package fake

import (
	"context"

	v1beta1 "github.com/kiegroup/kogito-operator/apis/app/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeKogitoBuilds implements KogitoBuildInterface
type FakeKogitoBuilds struct {
	Fake *FakeAppV1beta1
	ns   string
}

var kogitobuildsResource = schema.GroupVersionResource{Group: "app.kiegroup.org", Version: "v1beta1", Resource: "kogitobuilds"}

var kogitobuildsKind = schema.GroupVersionKind{Group: "app.kiegroup.org", Version: "v1beta1", Kind: "KogitoBuild"}

// Get takes name of the kogitoBuild, and returns the corresponding kogitoBuild object, and an error if there is any.
func (c *FakeKogitoBuilds) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.KogitoBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(kogitobuildsResource, c.ns, name), &v1beta1.KogitoBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.KogitoBuild), err
}

// List takes label and field selectors, and returns the list of KogitoBuilds that match those selectors.
func (c *FakeKogitoBuilds) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.KogitoBuildList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(kogitobuildsResource, kogitobuildsKind, c.ns, opts), &v1beta1.KogitoBuildList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.KogitoBuildList{ListMeta: obj.(*v1beta1.KogitoBuildList).ListMeta}
	for _, item := range obj.(*v1beta1.KogitoBuildList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested kogitoBuilds.
func (c *FakeKogitoBuilds) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(kogitobuildsResource, c.ns, opts))

}

// Create takes the representation of a kogitoBuild and creates it.  Returns the server's representation of the kogitoBuild, and an error, if there is any.
func (c *FakeKogitoBuilds) Create(ctx context.Context, kogitoBuild *v1beta1.KogitoBuild, opts v1.CreateOptions) (result *v1beta1.KogitoBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(kogitobuildsResource, c.ns, kogitoBuild), &v1beta1.KogitoBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.KogitoBuild), err
}

// Update takes the representation of a kogitoBuild and updates it. Returns the server's representation of the kogitoBuild, and an error, if there is any.
func (c *FakeKogitoBuilds) Update(ctx context.Context, kogitoBuild *v1beta1.KogitoBuild, opts v1.UpdateOptions) (result *v1beta1.KogitoBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(kogitobuildsResource, c.ns, kogitoBuild), &v1beta1.KogitoBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.KogitoBuild), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeKogitoBuilds) UpdateStatus(ctx context.Context, kogitoBuild *v1beta1.KogitoBuild, opts v1.UpdateOptions) (*v1beta1.KogitoBuild, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(kogitobuildsResource, "status", c.ns, kogitoBuild), &v1beta1.KogitoBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.KogitoBuild), err
}

// Delete takes name of the kogitoBuild and deletes it. Returns an error if one occurs.
func (c *FakeKogitoBuilds) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(kogitobuildsResource, c.ns, name), &v1beta1.KogitoBuild{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeKogitoBuilds) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(kogitobuildsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta1.KogitoBuildList{})
	return err
}

// Patch applies the patch and returns the patched kogitoBuild.
func (c *FakeKogitoBuilds) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.KogitoBuild, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(kogitobuildsResource, c.ns, name, pt, data, subresources...), &v1beta1.KogitoBuild{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.KogitoBuild), err
}
