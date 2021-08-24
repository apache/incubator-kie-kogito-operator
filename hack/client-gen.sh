#!/bin/bash
# Copyright 2021 Red Hat, Inc. and/or its affiliates
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

script_dir_path=`dirname "${BASH_SOURCE[0]}"`

set -o pipefail
set -o errtrace

location=$(dirname "$0")
# always consider to upgrade this version when upgrading Operator SDK or k8s/api
generator_version=v0.21.1
# in the future this can be an array
kogito_api_version=v1beta1

cd "$location"/../client || exit

error_handler() {
  if [ "$1" != "0" ]; then
    echo "$2"
    clean_up
    exit 1
  fi
}

clean_up() {
  rm -rf github.com
}

set_up() {
  rm -rf clientset
  rm -rf informers
  rm -rf listers
  go mod download k8s.io/code-generator
  error_handler $? "Failed to add k8s.io/code-generator to mod"
  go get k8s.io/code-generator/cmd/client-gen/generators@"${generator_version}"
  error_handler $? "Failed to add k8s.io/code-generator to mod"
}

gen_client() {
  go run k8s.io/code-generator/cmd/client-gen \
    --go-header-file=${script_dir_path}/boilerplate.go.txt \
    --input=/"${kogito_api_version}" \
    --input-base=github.com/kiegroup/kogito-operator/api \
    --input-dirs=github.com/kiegroup/kogito-operator/api/${kogito_api_version} \
    --clientset-name "versioned" \
    --output-base=. \
    --output-package=github.com/kiegroup/kogito-operator/client/clientset
  error_handler $? "Failed to generate clientset"
  move_gen_code
}

gen_listers() {
  go run k8s.io/code-generator/cmd/lister-gen \
    --go-header-file=${script_dir_path}/boilerplate.go.txt \
    --input-dirs=github.com/kiegroup/kogito-operator/api/${kogito_api_version} \
    --output-base=. \
    --output-package=github.com/kiegroup/kogito-operator/client/listers
  error_handler $? "Failed to generate listers"
}

gen_informers() {
  go run k8s.io/code-generator/cmd/informer-gen \
    --go-header-file=${script_dir_path}/boilerplate.go.txt \
    --versioned-clientset-package=github.com/kiegroup/kogito-operator/client/clientset/versioned \
    --internal-clientset-package=github.com/kiegroup/kogito-operator/client/clientset/versioned \
    --listers-package=github.com/kiegroup/kogito-operator/client/listers \
    --input-dirs=github.com/kiegroup/kogito-operator/api/${kogito_api_version} \
    --output-base=. \
    --output-package=github.com/kiegroup/kogito-operator/client/informers
  error_handler $? "Failed to generate informers"
}

move_gen_code() {
  mv github.com/kiegroup/kogito-operator/client/* .
  error_handler $? "Failed to move generated code to /client directory"
  if [ ! -f "clientset/versioned/typed/${kogito_api_version}/kogito_client.go" ]; then
    mv clientset/versioned/typed/${kogito_api_version}/_client.go clientset/versioned/typed/${kogito_api_version}/kogito_client.go
    error_handler $? "Failed to rename _client.go to kogito_client.go"
  fi
  if [ ! -f "clientset/versioned/typed/${kogito_api_version}/fake/fake_kogito_client.go" ]; then
    mv clientset/versioned/typed/${kogito_api_version}/fake/fake__client.go clientset/versioned/typed/${kogito_api_version}/fake/fake_kogito_client.go
    error_handler $? "Failed to rename fake__client.go to fake_kogito_client.go"
  fi
}

set_up
gen_client
gen_listers
gen_informers
move_gen_code
clean_up
