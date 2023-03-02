#!/bin/bash
# Copyright 2023 Red Hat, Inc. and/or its affiliates
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

VERSION=$1
if [[ $1 == v* ]]; then TAG=$1; else TAG=v$1; fi
GITHUB_AUTHOR=$2
COMMUNITY_OPERATORS=community-operators
COMMUNITY_OPERATORS_OWNER=k8s-operatorhub
COMMUNITY_OPERATORS_PROD=community-operators-prod
COMMUNITY_OPERATORS_PROD_OWNER=redhat-openshift-ecosystem
if [[ $3 == false ]]; then DRY_RUN=false; else DRY_RUN=true; fi
if [ -z "$4" ]; then KOGITO_OPERATOR_DIR=$(pwd); else KOGITO_OPERATOR_DIR=$4; fi

upstream_author=${UPSTREAM_AUTHOR:-kiegroup}

if [ "$GITHUB_AUTHOR" == "$upstream_author" ]; then
   echo "Upstream author and GitHub author are equal, no need to setup upstream ${upstream_author}"
else
    echo "Upstream author and GitHub author are equal. Adding upstream repo"
    git remote add upstream https://github.com/${upstream_author}/kogito-operator.git >/dev/null 2>&1
fi

echo "Kogito Operator directory is ${KOGITO_OPERATOR_DIR}"

git fetch --all --tags
echo "Checking out Kogito $TAG"
git checkout tags/$TAG

TMP_WD=$(mktemp -d)
echo "### Using the ${TMP_WD} tmp working dir ####"
create_operatorhub_pr() {
  cd $TMP_WD

  echo "### Starting changes on $1 repo ####"
  REPO_TO_CLONE="https://github.com/${GITHUB_AUTHOR}/$1"
  echo "$1 directory does not exist, going to clone ${REPO_TO_CLONE}"
  git clone ${REPO_TO_CLONE}

  cd $1

  git checkout -B kogito-$TAG
  KOGITO_TAG_INNER_DIR=operators/kogito-operator/${VERSION}
  echo "Creating inner dir ${KOGITO_TAG_INNER_DIR}"
  mkdir -p ${KOGITO_TAG_INNER_DIR}
  cp -rf ${KOGITO_OPERATOR_DIR}/bundle/app/* ${KOGITO_TAG_INNER_DIR}
  cp -f  ${KOGITO_OPERATOR_DIR}/bundle.Dockerfile ${KOGITO_TAG_INNER_DIR}/Dockerfile
  sed -i "s|bundle/app/manifests|manifests|g" ${KOGITO_TAG_INNER_DIR}/Dockerfile
  sed -i "s|bundle/app/metadata|metadata|g" ${KOGITO_TAG_INNER_DIR}/Dockerfile
  sed -i "s|bundle/app/tests|tests|g" ${KOGITO_TAG_INNER_DIR}/Dockerfile
  git add --all
  git commit --signoff -m "operator kogito-operator (${TAG})"
  if [[ ${DRY_RUN} == false ]]; then
    echo "We are running in non dry_run mode, going to push changes"
    if [[ -z ${GITHUB_TOKEN}  ]]; then
           echo "The GITHUB_TOKEN is not set, we will use SSH"
           ORIGIN_SSH="git@github.com:${GITHUB_AUTHOR}/${1}.git"
           git remote set-url origin ${ORIGIN_SSH}
    else
          echo "The GITHUB_TOKEN is set, we will use HTTPS"
          ORIGIN_HTTPS="https://${GITHUB_TOKEN}@github.com/${GITHUB_AUTHOR}/${1}.git"
          git remote set-url origin ${ORIGIN_HTTPS}
    fi
    git push --set-upstream origin kogito-$TAG
    if ! command -v gh &> /dev/null
    then
        echo "gh could not be found, you have to manually open a PR"
    else
      echo "We have found gh, going to open a draft PR"
      gh repo set-default $2/$1
      gh pr create --fill --body-file docs/pull_request_template.md --draft --base main
    fi
  fi
  echo "### Changes on $1 repo finished ####"
}

create_operatorhub_pr $COMMUNITY_OPERATORS $COMMUNITY_OPERATORS_OWNER
create_operatorhub_pr $COMMUNITY_OPERATORS_PROD $COMMUNITY_OPERATORS_PROD_OWNER

