// Copyright 2020 Red Hat, Inc. and/or its affiliates
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
	"archive/tar"
	"compress/gzip"
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestProduceTGZfile(t *testing.T) {
	tempDir := "/tmp/kogito-test/"
	simpleContent := []byte("hello World!!")
	files := []string{"file.bpmn", "file2.drl", "file3.bpmn2", "file4.dmn", "file5.properties", "file6.unsupported"}
	_, err := os.Stat(tempDir)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(tempDir, 0755)
		if errDir != nil {
			panic(err)
		}
	}
	for _, file := range files {
		err := ioutil.WriteFile(tempDir+file, simpleContent, 0755)
		fileInfo, err := os.Stat(tempDir + file)
		// double check file exists
		assert.Nil(t, err)
		assert.True(t, fileInfo.Size() > 0)
	}

	ioR, err := ProduceTGZfile(tempDir)
	assert.Nil(t, err)
	fileToWrite, err1 := ioutil.TempFile(os.TempDir(), "compressed_kogito_resources_*.tgz")
	assert.Nil(t, err1)

	if _, err2 := io.Copy(fileToWrite, ioR); err2 != nil {
		panic(err2)
	}
	defer fileToWrite.Close()

	// test if file is tgz and if the tgz contains the right files.
	// is Gzip?
	testGzip, err := os.Open(fileToWrite.Name())
	assert.Nil(t, err)
	defer testGzip.Close()

	// see https://mimesniff.spec.whatwg.org/
	// gzip 3 first bytes should be 1F 8B 08
	buff := make([]byte, 3)
	_, err = testGzip.Read(buff)
	assert.Nil(t, err)
	assert.Equal(t, hex.Dump(buff), "00000000  1f 8b 08                                          |...|\n")

	// extract file and see if it is tarball file
	testTar, err := os.Open(fileToWrite.Name())
	assert.Nil(t, err)

	gzf, err := gzip.NewReader(testTar)
	assert.Nil(t, err)

	tarHeader := tar.NewReader(gzf)

	var filesFromGzip []string
	for true {
		header, err := tarHeader.Next()
		if err == io.EOF {
			break
		}
		assert.Equal(t, header.Format, tar.FormatPAX)
		assert.NotNil(t, header.Name)
		filesFromGzip = append(filesFromGzip, header.Name)
	}

	// all files extracted from gzip must be on the initial slice of test files
	for _, fileFromGzip := range filesFromGzip {
		assert.True(t, contains(fileFromGzip, files))
	}

	// And filesFromGzip should not contain the unsupported file: file6.unsupported
	assert.False(t, contains("file6.unsupported", filesFromGzip))

	os.RemoveAll(tempDir)
	os.Remove(fileToWrite.Name())
}

func TestIsSuffixSupported(t *testing.T) {
	//test a few invalid extensions
	invalidExtensions := []string{"test.dn", "test.st", "test.www", "test.bmn2", "test.prop"}
	for _, ext := range invalidExtensions {
		assert.False(t, IsSuffixSupported(ext))
	}

	//test a few valid extensions
	validExtensions := []string{"test.drl", "test.dmn", "test.bpmn2", "test.properties", "test.bpmn"}
	for _, ext := range validExtensions {
		assert.True(t, IsSuffixSupported(ext))
	}
}

func contains(item string, slice []string) bool {
	for _, value := range slice {
		if value == item {
			return true
		}
	}
	return false
}
