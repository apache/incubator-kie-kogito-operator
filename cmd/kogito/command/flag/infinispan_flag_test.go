// Copyright v Red Hat, Inc. and/or its affiliates
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

package flag

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CheckInfinispanArgs(t *testing.T) {

	// Scenario 1 : Username provide and password not provided
	flags := &InfinispanFlags{InfinispanUser: "username"}
	err := CheckInfinispanArgs(flags)
	assert.NotNil(t, err)

	// Scenario 2 : Username not provide and password provided
	flags = &InfinispanFlags{InfinispanPassword: "password"}
	err = CheckInfinispanArgs(flags)
	assert.NotNil(t, err)

	// Scenario 3 : URI provided and Username/password not provided
	flags = &InfinispanFlags{URI: "uri"}
	err = CheckInfinispanArgs(flags)
	assert.NotNil(t, err)

	// Scenario 4 : URI provided and Username/password is also provided
	flags = &InfinispanFlags{URI: "uri", InfinispanUser: "username", InfinispanPassword: "password"}
	err = CheckInfinispanArgs(flags)
	assert.Nil(t, err)
}
