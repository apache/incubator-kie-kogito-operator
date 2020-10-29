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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetKogitoServiceInternalURL(t *testing.T) {
	service := &v1alpha1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dataindex",
			Namespace: "mynamespace",
		},
		Spec: v1alpha1.KogitoSupportingServiceSpec{
			ServiceType: v1alpha1.DataIndex,
		},
	}

	actualURL := GetKogitoServiceURL(service)
	assert.Equal(t, "http://dataindex.mynamespace", actualURL)
}

func Test_getKogitoDataIndexURLs(t *testing.T) {
	ns := t.Name()
	hostname := "dataindex-route.com"
	expectedHTTPURL := "http://" + hostname
	expectedWSURL := "ws://" + hostname
	expectedHTTPSURL := "https://" + hostname
	expectedWSSURL := "wss://" + hostname
	insecureDI := &v1alpha1.KogitoSupportingService{
		Spec: v1alpha1.KogitoSupportingServiceSpec{
			ServiceType: v1alpha1.DataIndex,
		},
		ObjectMeta: metav1.ObjectMeta{Name: DefaultDataIndexName, Namespace: ns},
		Status:     v1alpha1.KogitoSupportingServiceStatus{KogitoServiceStatus: v1alpha1.KogitoServiceStatus{ExternalURI: expectedHTTPURL}},
	}
	secureDI := &v1alpha1.KogitoSupportingService{
		Spec: v1alpha1.KogitoSupportingServiceSpec{
			ServiceType: v1alpha1.DataIndex,
		},
		ObjectMeta: metav1.ObjectMeta{Name: DefaultDataIndexName, Namespace: ns},
		Status:     v1alpha1.KogitoSupportingServiceStatus{KogitoServiceStatus: v1alpha1.KogitoServiceStatus{ExternalURI: expectedHTTPSURL}},
	}
	cliInsecure := test.NewFakeClientBuilder().AddK8sObjects(insecureDI).Build()
	cliSecure := test.NewFakeClientBuilder().AddK8sObjects(secureDI).Build()
	type args struct {
		client    *client.Client
		namespace string
	}
	tests := []struct {
		name        string
		args        args
		wantHTTPURL string
		wantWSURL   string
		wantErr     bool
	}{
		{
			name: "With insecure route",
			args: args{
				client:    cliInsecure,
				namespace: ns,
			},
			wantHTTPURL: expectedHTTPURL,
			wantWSURL:   expectedWSURL,
			wantErr:     false,
		},
		{
			name: "With secure route",
			args: args{
				client:    cliSecure,
				namespace: ns,
			},
			wantHTTPURL: expectedHTTPSURL,
			wantWSURL:   expectedWSSURL,
			wantErr:     false,
		},
		{
			name: "With blank route",
			args: args{
				client:    test.NewFakeClientBuilder().Build(),
				namespace: ns,
			},
			wantHTTPURL: "",
			wantWSURL:   "",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDataIndexEndpoints, err := getServiceEndpoints(tt.args.client, tt.args.namespace, tt.wantHTTPURL, tt.wantWSURL, v1alpha1.DataIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDataIndexEndpoints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotDataIndexEndpoints != nil &&
				gotDataIndexEndpoints.HTTPRouteURI != tt.wantHTTPURL {
				t.Errorf("GetDataIndexEndpoints() gotHTTPURL = %v, want %v", gotDataIndexEndpoints.HTTPRouteURI, tt.wantHTTPURL)
			}
			if gotDataIndexEndpoints != nil &&
				gotDataIndexEndpoints.WSRouteURI != tt.wantWSURL {
				t.Errorf("GetDataIndexEndpoints() gotWSURL = %v, want %v", gotDataIndexEndpoints.WSRouteURI, tt.wantWSURL)
			}
		})
	}
}
