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


if [[ -z ${ENVTEST_ASSETS_DIR} ]]; then
  ENVTEST_ASSETS_DIR=testbin
fi

testdir=$(pwd)/${ENVTEST_ASSETS_DIR}
mkdir -p ${testdir}
test -f  ${testdir}/setup-envtest.sh || curl -sSLo ${testdir}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.6.3/hack/setup-envtest.sh
sed -i "s,#\!.*,#\!\/bin\/bash,g"  ${testdir}/setup-envtest.sh
source  ${testdir}/setup-envtest.sh; fetch_envtest_tools  ${testdir}; setup_envtest_env  ${testdir}; \
echo "Testing root"
go test ./cmd/... -p=1 -count=1 -coverprofile cmd-cover.out; \
go test ./controllers/... -p=1 -count=1 -coverprofile controllers-cover.out
go test ./core/... -p=1 -count=1 -coverprofile core-cover.out
echo "Testing apis"
cd apis && go test ./... -p=1 -count=1 -coverprofile apis-cover.out
cd - || exit
echo "Testing client"
cd client && go test ./... -p=1 -count=1 -coverprofile client-cover.out
cd - || exit