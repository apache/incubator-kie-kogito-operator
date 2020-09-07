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


./hack/run-tests.sh \
	--feature scripts/examples \
	--tags "~@native" \
	--concurrent 2 \
	--timeout 240 \
	--ci GHActions \
	--operator_image localhost:5000/kiegroup/kogito-cloud-operator \
	--operator_tag latest \
	--runtime_application_image_registry localhost:5000 \
	--runtime_application_image_namespace kiegroup \
	--runtime_application_image_version latest \
	--load_factor 2 \
	--container_engine docker \
	--domain_suffix example.com  \
	--cr_deployment_only \
	--load_default_config
