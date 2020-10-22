// Copyright 2020 Red Hat, Inc. and/or its affiliates
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

package infrastructure

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"knative.dev/eventing/pkg/apis/eventing"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
)

const (
	// KnativeEventingBrokerKind is the Kind description for Knative Eventing Brokers
	KnativeEventingBrokerKind = "Broker"
)

var (
	// KnativeEventingAPIVersion API Group version as defined by Knative Eventing operator
	KnativeEventingAPIVersion = eventingv1.SchemeGroupVersion.String()
)

// IsKnativeEventingAvailable checks if Knative Eventing CRDs are available in the cluster
func IsKnativeEventingAvailable(client *client.Client) bool {
	return client.HasServerGroup(eventing.GroupName)
}
