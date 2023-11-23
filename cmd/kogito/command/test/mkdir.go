// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package test

import (
	"fmt"
	"io/ioutil"
	"os"
)

// Mkdir creates directory dir, including non-existing parent directories.
func Mkdir(dir string) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dir, 0755)
		if errDir != nil {
			panic(err)
		}
	}
}

// TempDirWithFile creates a temporary directory dir with a
// temporary file
func TempDirWithFile(dir string, file string) string {
	tmpDir, err := ioutil.TempDir("", dir)
	if err != nil {
		panic(err)
	}
	_, err = ioutil.TempFile(tmpDir, file)
	if err != nil {
		panic(err)
	}
	return tmpDir
}

// TempDirWithSubDir creates a temporary directory dir with a
// temporary subDir
func TempDirWithSubDir(dir string, subDir string) string {
	tmpDir, err := ioutil.TempDir("", dir)
	if err != nil {
		panic(err)
	}
	err = os.Mkdir(fmt.Sprintf("%s/%s", tmpDir, subDir), 0755)
	if err != nil {
		panic(err)
	}
	return tmpDir
}
