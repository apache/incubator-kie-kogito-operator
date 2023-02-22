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
COMMUNITY_OPERATORS_DIR=community-operators
COMMUNITY_OPERATORS_REPO=$2
COMMUNITY_OPERATORS_PROD_DIR=community-operators-prod
COMMUNITY_OPERATORS_PROD_REPO=$3
if [[ $4 == false ]]; then DRY_RUN=false; else DRY_RUN=true; fi

git fetch --tags --all
echo "Checking out Kogito $TAG"
git checkout tags/$TAG -B kogito-$TAG

cd ../

create_operatorhub_pr() {
  if [ -d "$1" ];
  then
      echo "$1 directory exists."
      cd $1
      git fetch --all
      git checkout main
      git merge upstream/main
  else
  	echo "$1 directory does not exist."
    git clone $2
    cd $1
  fi


  ls
  git checkout -B kogito-$TAG
  cd operators/kogito-operator
  mkdir -p ${VERSION}
  cd ${VERSION}
  ls
  cp -rf ../../../../kogito-operator/bundle/app/* .
  cp -f ../../../../kogito-operator/bundle.Dockerfile Dockerfile
  sed -i "s|bundle/app/manifests|manifests|g" Dockerfile
  sed -i "s|bundle/app/metadata|metadata|g" Dockerfile
  sed -i "s|bundle/app/tests|tests|g" Dockerfile
  git add .
  git commit --signoff -m  "operator kogito-operator (${TAG})"
  if [[ ${DRY_RUN} == false ]]; then git push -uf; fi
  cd ../../../../
}

create_operatorhub_pr $COMMUNITY_OPERATORS_DIR $COMMUNITY_OPERATORS_REPO
create_operatorhub_pr $COMMUNITY_OPERATORS_PROD_DIR $COMMUNITY_OPERATORS_PROD_REPO

