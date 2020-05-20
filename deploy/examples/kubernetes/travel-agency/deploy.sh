#!/bin/env bash

EXAMPLES_DIR=./deploy/examples
TARGET_DIR="${EXAMPLES_DIR}/kubernetes/travel-agency"
STRIMZI_VERSION=0.17.0
if [[ -z ${PROJECT_NAME} ]]; then
  PROJECT_NAME=kogito
fi

echo "Deploying the Travel Agency Demo on ${PROJECT_NAME} namespace"
oc project "$PROJECT_NAME"

echo "Installing Infinispan Operator"
oc apply -f https://raw.githubusercontent.com/infinispan/infinispan-operator/1.1.1.Final/deploy/crd.yaml -n ${PROJECT_NAME}
oc apply -f https://raw.githubusercontent.com/infinispan/infinispan-operator/1.1.1.Final/deploy/rbac.yaml -n ${PROJECT_NAME}
oc apply -f https://raw.githubusercontent.com/infinispan/infinispan-operator/1.1.1.Final/deploy/operator.yaml -n ${PROJECT_NAME}

echo "Deploying Strimzi"
wget "https://github.com/strimzi/strimzi-kafka-operator/releases/download/${STRIMZI_VERSION}/strimzi-${STRIMZI_VERSION}.tar.gz" -P "$TARGET_DIR/"
tar zxf "${TARGET_DIR}/strimzi-${STRIMZI_VERSION}.tar.gz" -C "$TARGET_DIR"
find ${TARGET_DIR}/strimzi-${STRIMZI_VERSION}/install/cluster-operator -name '*RoleBinding*.yaml' -type f -exec sed -i "s/namespace: .*/namespace: ${PROJECT_NAME}/" {} \;
oc apply -f ${TARGET_DIR}/strimzi-${STRIMZI_VERSION}/install/cluster-operator/ -n ${PROJECT_NAME}

echo "Deploying Data Index"
oc apply -f ${TARGET_DIR}/data-index.yaml -n ${PROJECT_NAME}
oc apply -f ${TARGET_DIR}/data-index-ingress.yaml -n ${PROJECT_NAME}

echo "Deploying Kogito Travels Application"
oc apply -f ${TARGET_DIR}/kogito-travels.yaml -n ${PROJECT_NAME}
oc apply -f ${TARGET_DIR}/kogito-travels-ingress.yaml -n ${PROJECT_NAME}
oc apply -f ${TARGET_DIR}/kogito-visas.yaml -n ${PROJECT_NAME}
oc apply -f ${TARGET_DIR}/kogito-visas-ingress.yaml -n ${PROJECT_NAME}
