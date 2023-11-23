#!/bin/env bash
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

EXAMPLES_DIR=./examples
TARGET_DIR="${EXAMPLES_DIR}/kubernetes/travel-agency"
STRIMZI_VERSION=0.17.0
INFINISPAN_VERSION=1.1.1.Final

if [[ -z ${PROJECT_NAME} ]]; then
  PROJECT_NAME=kogito
fi

echo "Deploying the Travel Agency Demo on ${PROJECT_NAME} namespace"
kubectl create namespace "$PROJECT_NAME"

echo "Installing Infinispan Operator"
kubectl apply -f "https://raw.githubusercontent.com/infinispan/infinispan-operator/${INFINISPAN_VERSION}/deploy/crd.yaml" -n ${PROJECT_NAME}
kubectl apply -f "https://raw.githubusercontent.com/infinispan/infinispan-operator/${INFINISPAN_VERSION}/deploy/rbac.yaml" -n ${PROJECT_NAME}
kubectl apply -f "https://raw.githubusercontent.com/infinispan/infinispan-operator/${INFINISPAN_VERSION}/deploy/operator.yaml" -n ${PROJECT_NAME}

echo "Deploying Strimzi"
wget "https://github.com/strimzi/strimzi-kafka-operator/releases/download/${STRIMZI_VERSION}/strimzi-${STRIMZI_VERSION}.tar.gz" -P "$TARGET_DIR/"
tar zxf "${TARGET_DIR}/strimzi-${STRIMZI_VERSION}.tar.gz" -C "$TARGET_DIR"
find ${TARGET_DIR}/strimzi-${STRIMZI_VERSION}/install/cluster-operator -name '*RoleBinding*.yaml' -type f -exec sed -i "s/namespace: .*/namespace: ${PROJECT_NAME}/" {} \;
kubectl apply -f ${TARGET_DIR}/strimzi-${STRIMZI_VERSION}/install/cluster-operator/ -n ${PROJECT_NAME}

echo "Deploying Data Index"
kubectl apply -f ${TARGET_DIR}/data-index.yaml -n ${PROJECT_NAME}
kubectl apply -f ${TARGET_DIR}/data-index-ingress.yaml -n ${PROJECT_NAME}

echo "Deploying Management Console"
kubectl apply -f ${TARGET_DIR}/management-console.yaml -n ${PROJECT_NAME}
kubectl apply -f ${TARGET_DIR}/management-console-ingress.yaml -n ${PROJECT_NAME}

echo "Deploying Kogito Travels Application"
kubectl apply -f ${TARGET_DIR}/kogito-travels.yaml -n ${PROJECT_NAME}
kubectl apply -f ${TARGET_DIR}/kogito-travels-ingress.yaml -n ${PROJECT_NAME}
kubectl apply -f ${TARGET_DIR}/kogito-visas.yaml -n ${PROJECT_NAME}
kubectl apply -f ${TARGET_DIR}/kogito-visas-ingress.yaml -n ${PROJECT_NAME}
