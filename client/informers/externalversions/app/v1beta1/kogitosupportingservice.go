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
// Code generated by informer-gen. DO NOT EDIT.

package v1beta1

import (
	"context"
	time "time"

	appv1beta1 "github.com/apache/incubator-kie-kogito-operator/apis/app/v1beta1"
	versioned "github.com/apache/incubator-kie-kogito-operator/client/clientset/versioned"
	internalinterfaces "github.com/apache/incubator-kie-kogito-operator/client/informers/externalversions/internalinterfaces"
	v1beta1 "github.com/apache/incubator-kie-kogito-operator/client/listers/app/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// KogitoSupportingServiceInformer provides access to a shared informer and lister for
// KogitoSupportingServices.
type KogitoSupportingServiceInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1beta1.KogitoSupportingServiceLister
}

type kogitoSupportingServiceInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewKogitoSupportingServiceInformer constructs a new informer for KogitoSupportingService type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewKogitoSupportingServiceInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredKogitoSupportingServiceInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredKogitoSupportingServiceInformer constructs a new informer for KogitoSupportingService type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredKogitoSupportingServiceInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AppV1beta1().KogitoSupportingServices(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AppV1beta1().KogitoSupportingServices(namespace).Watch(context.TODO(), options)
			},
		},
		&appv1beta1.KogitoSupportingService{},
		resyncPeriod,
		indexers,
	)
}

func (f *kogitoSupportingServiceInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredKogitoSupportingServiceInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *kogitoSupportingServiceInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&appv1beta1.KogitoSupportingService{}, f.defaultInformer)
}

func (f *kogitoSupportingServiceInformer) Lister() v1beta1.KogitoSupportingServiceLister {
	return v1beta1.NewKogitoSupportingServiceLister(f.Informer().GetIndexer())
}
