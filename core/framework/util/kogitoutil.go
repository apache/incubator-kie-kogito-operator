// Copyright 2021 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"os"
	"strings"
)

// IsProductMode returns true if application is running in product mode.
func IsProductMode() bool {
	group, _ := os.LookupEnv("GROUP")
	return strings.ToUpper(group) == "RHPAM"
}

// IsDebugMode returns true if application is running in debug mode.
func IsDebugMode() bool {
	debug, _ := os.LookupEnv("DEBUG")
	return strings.ToUpper(debug) == "TRUE"
}
