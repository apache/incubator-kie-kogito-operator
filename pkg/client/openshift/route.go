// Copyright 2019 Red Hat, Inc. and/or its affiliates
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

package openshift

import (
	"time"

	routev1 "github.com/openshift/api/route/v1"

	"k8s.io/apimachinery/pkg/types"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
)

// RouteInterface exposes common operations for Route API
type RouteInterface interface {
	GetHostFromRoute(routeKey types.NamespacedName) (string, error)
}

func newRoute(c *client.Client) RouteInterface {
	client.MustEnsureClient(c)
	return &route{
		client: c,
	}
}

type route struct {
	client *client.Client
}

func (r *route) GetHostFromRoute(routeKey types.NamespacedName) (string, error) {
	route := &routev1.Route{}

	for i := 1; i < 60; i++ {
		time.Sleep(time.Duration(100) * time.Millisecond)
		if exists, err :=
			kubernetes.ResourceC(r.client).FetchWithKey(routeKey, route); exists {
			break
		} else if err != nil {
			log.Error("Error getting Route. ", err)
			return "", err
		}
	}

	return route.Spec.Host, nil
}
