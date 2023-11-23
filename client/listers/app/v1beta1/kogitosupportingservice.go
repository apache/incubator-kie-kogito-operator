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
// Code generated by lister-gen. DO NOT EDIT.

package v1beta1

import (
	v1beta1 "github.com/kiegroup/kogito-operator/apis/app/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// KogitoSupportingServiceLister helps list KogitoSupportingServices.
// All objects returned here must be treated as read-only.
type KogitoSupportingServiceLister interface {
	// List lists all KogitoSupportingServices in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.KogitoSupportingService, err error)
	// KogitoSupportingServices returns an object that can list and get KogitoSupportingServices.
	KogitoSupportingServices(namespace string) KogitoSupportingServiceNamespaceLister
	KogitoSupportingServiceListerExpansion
}

// kogitoSupportingServiceLister implements the KogitoSupportingServiceLister interface.
type kogitoSupportingServiceLister struct {
	indexer cache.Indexer
}

// NewKogitoSupportingServiceLister returns a new KogitoSupportingServiceLister.
func NewKogitoSupportingServiceLister(indexer cache.Indexer) KogitoSupportingServiceLister {
	return &kogitoSupportingServiceLister{indexer: indexer}
}

// List lists all KogitoSupportingServices in the indexer.
func (s *kogitoSupportingServiceLister) List(selector labels.Selector) (ret []*v1beta1.KogitoSupportingService, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.KogitoSupportingService))
	})
	return ret, err
}

// KogitoSupportingServices returns an object that can list and get KogitoSupportingServices.
func (s *kogitoSupportingServiceLister) KogitoSupportingServices(namespace string) KogitoSupportingServiceNamespaceLister {
	return kogitoSupportingServiceNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// KogitoSupportingServiceNamespaceLister helps list and get KogitoSupportingServices.
// All objects returned here must be treated as read-only.
type KogitoSupportingServiceNamespaceLister interface {
	// List lists all KogitoSupportingServices in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.KogitoSupportingService, err error)
	// Get retrieves the KogitoSupportingService from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1beta1.KogitoSupportingService, error)
	KogitoSupportingServiceNamespaceListerExpansion
}

// kogitoSupportingServiceNamespaceLister implements the KogitoSupportingServiceNamespaceLister
// interface.
type kogitoSupportingServiceNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all KogitoSupportingServices in the indexer for a given namespace.
func (s kogitoSupportingServiceNamespaceLister) List(selector labels.Selector) (ret []*v1beta1.KogitoSupportingService, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.KogitoSupportingService))
	})
	return ret, err
}

// Get retrieves the KogitoSupportingService from the indexer for a given namespace and name.
func (s kogitoSupportingServiceNamespaceLister) Get(name string) (*v1beta1.KogitoSupportingService, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1beta1.Resource("kogitosupportingservice"), name)
	}
	return obj.(*v1beta1.KogitoSupportingService), nil
}
