/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package test

import "os"

// SetSharedEnv sets a value to a given Environment variable
// returns the defer function that MUST be called after your test to not mess up with users' env
func SetSharedEnv(k, v string) (deferFunc func()) {
	backupValue := os.Getenv(k)
	_ = os.Setenv(k, v)
	return func() {
		if len(backupValue) > 0 {
			_ = os.Setenv(k, backupValue)
		}
	}
}
