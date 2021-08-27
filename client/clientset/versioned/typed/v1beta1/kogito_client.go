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
	v1beta1 "github.com/kiegroup/kogito-operator/api/v1beta1"
	"github.com/kiegroup/kogito-operator/client/clientset/versioned/scheme"
	rest "k8s.io/client-go/rest"
)

type V1beta1Interface interface {
	RESTClient() rest.Interface
	KogitoBuildsGetter
	KogitoInfrasGetter
	KogitoRuntimesGetter
	KogitoSupportingServicesGetter
}

// V1beta1Client is used to interact with features provided by the  group.
type V1beta1Client struct {
	restClient rest.Interface
}

func (c *V1beta1Client) KogitoBuilds(namespace string) KogitoBuildInterface {
	return newKogitoBuilds(c, namespace)
}

func (c *V1beta1Client) KogitoInfras(namespace string) KogitoInfraInterface {
	return newKogitoInfras(c, namespace)
}

func (c *V1beta1Client) KogitoRuntimes(namespace string) KogitoRuntimeInterface {
	return newKogitoRuntimes(c, namespace)
}

func (c *V1beta1Client) KogitoSupportingServices(namespace string) KogitoSupportingServiceInterface {
	return newKogitoSupportingServices(c, namespace)
}

// NewForConfig creates a new V1beta1Client for the given config.
func NewForConfig(c *rest.Config) (*V1beta1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &V1beta1Client{client}, nil
}

// NewForConfigOrDie creates a new V1beta1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *V1beta1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new V1beta1Client for the given RESTClient.
func New(c rest.Interface) *V1beta1Client {
	return &V1beta1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1beta1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *V1beta1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
