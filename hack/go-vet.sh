#!/bin/bash
# Copyright 2019 Red Hat, Inc. and/or its affiliates
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


./hack/go-mod.sh
# enforce GOROOT
export GOROOT=$(go env GOROOT)
export GOPATH=$(go env GOPATH)
operator-sdk generate k8s
operator-sdk generate crds --crd-version=v1beta1
# get the openapi binary
which openapi-gen > /dev/null || go build -o $GOPATH/bin/openapi-gen k8s.io/kube-openapi/cmd/openapi-gen
# generate the openapi files
echo "Generating openapi files"
openapi-gen --logtostderr=true -o "" -i github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1 -O zz_generated.openapi -p ./pkg/apis/app/v1alpha1 -h ./hack/boilerplate.go.txt -r "-"
openapi-gen --logtostderr=true -o "" -i github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1 -O zz_generated.openapi -p ./pkg/apis/kafka/v1beta1 -h ./hack/boilerplate.go.txt -r "-"

./hack/generate-manifests.sh

go vet ./...
