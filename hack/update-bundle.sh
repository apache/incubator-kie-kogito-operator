#!/bin/bash
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.


set -e


# The script will update the containerImage inside CSV and remove the replaces field
if [[ -z ${1} ]]; then
    echo "No image given. Please provide the image to use in CSV"
    exit 1
fi

echo "Will update the bundle CSV with ${1} image"

## update the image
sed -i  "s|quay.io/kiegroup/kogito-operator.*|${1}|g" bundle/app/manifests/kogito-operator.clusterserviceversion.yaml

## remove the replaces field
sed -i "/replaces.*/d" bundle/app/manifests/kogito-operator.clusterserviceversion.yaml

echo "Bundle CSV updated"
