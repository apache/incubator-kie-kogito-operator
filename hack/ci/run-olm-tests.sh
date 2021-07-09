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

# Requirements to run OLM tests:
# 1. Have docker installed on your system
# 2. Passwordless sudo is required by the OLM script to run tests
# 3. Have ansible installed on your system
# That's it all other tools will be downloaded and installed by the OLM script.

set -e
source ./hack/ci/operator-ensure-manifests.sh

# SCRIPT_URL URL to the script used by OLM to test the operator
SCRIPT_URL="https://raw.githubusercontent.com/operator-framework/operator-test-playbooks/master/upstream/test/test.sh"

cd "${tempfolder}"

bash <(curl -sL "${SCRIPT_URL}") all community-operators/kogito-operator/"${version}"
#bash <(curl -sL "${SCRIPT_URL}") docker all community-operators/kogito-operator/"${version}"