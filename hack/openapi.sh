#!/bin/bash
# Copyright 2020 Red Hat, Inc. and/or its affiliates
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


source ./hack/go-path.sh

# get the openapi binary
command -v openapi-gen >/dev/null || go build -o "${GOPATH}"/bin/openapi-gen k8s.io/kube-openapi/cmd/openapi-gen
echo "Generating openapi files"
openapi-gen --logtostderr=true -o "" -i ./api/v1beta1 -O zz_generated.openapi -p ./api/v1beta1 -h ./hack/boilerplate.go.txt -r "-"
openapi-gen --logtostderr=true -o "" -i ./api/kafka/v1beta1 -O zz_generated.openapi -p ./api/kafka/v1beta1 -h ./hack/boilerplate.go.txt -r "-"
