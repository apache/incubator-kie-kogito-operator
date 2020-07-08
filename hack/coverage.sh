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


# The script accepts the coverage report output file as argument
# Which is generated by the go test

if [ -z "$1" ]
  then
    echo "No argument supplied"
    exit 1
fi

coverage=$(go tool cover -func $1 | grep total | awk '{print substr($3, 1, length($3)-1)}')

result=$( bc <<< "${coverage%G} < $MIN_COVERAGE" )

if [[ $result == 1 ]]; then
  echo "Coverage is $coverage, which is less than the required minimum coverage: $MIN_COVERAGE "
  code=1
fi

go tool cover -html=$1 -o $1
echo "Coverage for $1 is $coverage"
echo "Please see the detailed coverage report in artifacts section in $1 file"

exit ${code:0}