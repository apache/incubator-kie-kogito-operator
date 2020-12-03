#!/bin/env bash
# Copyright 2019 Red Hat, Inc. and/or its affiliates
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

declare exit_status=0
declare file_found=0

shopt -s nullglob
for yaml in deploy/crds/*_crd.yaml; do
  file_found=1
  kubectl apply -f "./${yaml}"
  exit_status=$?
  if [ $exit_status -gt 0 ]; then
    break # Don't try other files if one fails
  fi
done
shopt -u nullglob

if [[ file_found -eq 0 ]]; then
  echo "No deployment files found" >&2
  exit_status=3
fi

exit ${exit_status}
