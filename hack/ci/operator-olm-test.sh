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

set -e
source ./hack/ci/operator-ensure-manifests.sh
source ./hack/export-version.sh

default_cluster_name="operator-test"

cluster_type=$1
container_engine="docker"
container_extra_args=""
CATALOG_IMAGE="operatorhubio-catalog:temp"
if [[ "${cluster_type}" = "openshift" ]]; then
    which oc > /dev/null || (echo "oc is not installed. Please install it before proceeding. Exiting...." && exit 1)
    which podman > /dev/null || (echo "podman is not installed. Please install it before proceeding. Exiting...." && exit 1)
    if [[ -z ${OPENSHIFT_REGISTRY} ]]; then
        echo "OPENSHIFT_REGISTRY is not defined, please define it"
        exit 1
    fi

    container_engine="podman"
    container_extra_args="--tls-verify=false"
    CATALOG_IMAGE="${OPENSHIFT_REGISTRY}/openshift/${CATALOG_IMAGE}"
else
    which docker > /dev/null || (echo "docker is not installed. Please install it before proceeding. Exiting...." && exit 1)
    which kind > /dev/null || (echo "kind is not installed. Please install it before proceeding. Exiting...." && exit 1)
    if [[ -z ${CLUSTER_NAME} ]]; then
        CLUSTER_NAME=$default_cluster_name
    fi

    cluster_type="kind"
fi

OP_PATH="community-operators/kogito-operator"
INSTALL_MODE="SingleNamespace"
# Changed image with older version of operator-courier
# See: https://github.com/operator-framework/operator-courier/issues/188
# Needs to be changed back to quay.io/operator-framework/operator-testing:latest
# after the issue has been resolved and the new image has been built
OPERATOR_TESTING_IMAGE="quay.io/operator-framework/operator-testing:latest"

if [ -z ${KUBECONFIG} ]; then
    KUBECONFIG=${HOME}/.kube/config
    echo "---> KUBECONFIG environment variable not set, defining to:"
    ls -la ${KUBECONFIG}
fi

echo "---> Loading Operator Image into ${cluster_type}"
if [[ "${cluster_type}" = "openshift" ]]; then
    sed -i "s|image:.*|image: image-registry.openshift-image-registry.svc:5000/openshift/kogito-cloud-operator:${OP_VERSION}|g" deploy/operator.yaml
    ./hack/generate-manifests.sh
    echo "${container_engine} login -u anything -p $(oc whoami -t) ${container_extra_args} ${OPENSHIFT_REGISTRY}"
    ${container_engine} login -u anything -p $(oc whoami -t) ${container_extra_args} ${OPENSHIFT_REGISTRY}
    echo "${container_engine} tag quay.io/kiegroup/kogito-cloud-operator:\"${OP_VERSION}\" \"${OPENSHIFT_REGISTRY}/openshift/kogito-cloud-operator:${OP_VERSION}\""
    ${container_engine} tag quay.io/kiegroup/kogito-cloud-operator:"${OP_VERSION}" "${OPENSHIFT_REGISTRY}/openshift/kogito-cloud-operator:${OP_VERSION}"
    echo "${container_engine} push ${container_extra_args} \"${OPENSHIFT_REGISTRY}/openshift/kogito-cloud-operator:${OP_VERSION}\""
    ${container_engine} push ${container_extra_args} "${OPENSHIFT_REGISTRY}/openshift/kogito-cloud-operator:${OP_VERSION}"
else 
    kind load docker-image quay.io/kiegroup/kogito-cloud-operator:"${OP_VERSION}" --name ${CLUSTER_NAME}
    node_name=$(kubectl get nodes -o jsonpath="{.items[0].metadata.name}")
    echo "---> Checking internal loaded images on node ${node_name}"
    ${container_engine} exec "${node_name}" crictl images
fi

echo "---> Updating All CSV files to imagePullPolicy: Never"
find "${OUTPUT}/kogito-operator/" -name '*.clusterserviceversion.yaml' -type f -print0 | xargs -0 sed -i 's/imagePullPolicy: Always/imagePullPolicy: Never/g'
echo "---> Resulting imagePullPolicy on manifest files"
grep -rn imagePullPolicy ${OUTPUT}/kogito-operator
echo "---> Building temporary catalog Image"
${container_engine} build --build-arg PERMISSIVE_LOAD=false -f ./hack/ci/operatorhubio-catalog.Dockerfile -t ${CATALOG_IMAGE} .
echo "---> Loading Catalog Image into ${cluster_type}"
if [[ "${cluster_type}" = "openshift" ]]; then
    ${container_engine} push ${container_extra_args} ${CATALOG_IMAGE}
else 
    kind load docker-image ${CATALOG_IMAGE} --name ${CLUSTER_NAME}
fi

# running tests
${container_engine} pull ${OPERATOR_TESTING_IMAGE}
${container_engine} run --network=host --rm \
    -v ${KUBECONFIG}:/root/.kube/config:z \
    -v ${OUTPUT}/:/community-operators:z ${OPERATOR_TESTING_IMAGE} \
    operator.test --no-print-directory \
    OP_PATH=${OP_PATH} VERBOSE=true NO_KIND=0 CATALOG_IMAGE=${CATALOG_IMAGE} INSTALL_MODE=${INSTALL_MODE}
