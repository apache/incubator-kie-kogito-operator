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

source ./hack/env.sh

CSV_DIR="config/manifests/bases/"

old_version=$(getOperatorVersion)
new_version=$1

if [ -z "$new_version" ]; then
  echo "Please inform the new version. Use X.X.X"
  exit 1
fi

sed -i "s/$old_version/$new_version/g" cmd/kogito/version/version.go README.md version/version.go deploy/operator.yaml ${OLM_DIR}/kogito-operator.package.yaml ${OLM_DIR}/custom-subscription-example.yaml hack/go-build.sh hack/go-vet.sh .osdk-scorecard.yaml

make vet


# replace in csv file
csv_file="${CSV_DIR}/kogito-operator.clusterserviceversion.yaml"
sed -i "s|replaces: kogito-operator.*|replaces: kogito-operator.v${LATEST_RELEASED_OLM_VERSION}|g" ${csv_file}
sed -i "s/$old_version/$new_version/g" ${csv_file}

make bundle

# rewrite test default config, all other configuration into the file will be overridden
./hack/update_test_config.sh


echo "Version bumped from $old_version to $new_version"
