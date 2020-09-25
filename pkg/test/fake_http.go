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

package test

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/restclient"
	"net/http"
)

func newFakeHTTPClient(fakeHTTPReqs map[*http.Request]DoFunc) restclient.HTTPClient {
	doFuncs := map[string]DoFunc{}
	for req, doFunc := range fakeHTTPReqs {
		doFuncs[createKey(req)] = doFunc
	}
	return &MockHTTPClient{
		doFunc: doFuncs,
	}
}

// DoFunc represent http do func implementation
type DoFunc func(*http.Request) (*http.Response, error)

// MockHTTPClient is the mock client
type MockHTTPClient struct {
	doFunc map[string]DoFunc
}

// Do is the mock client's `Do` func
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	doFunc := m.doFunc[createKey(req)]
	return doFunc(req)
}

func createKey(req *http.Request) string {
	return req.Method + ":" + req.URL.String()
}
