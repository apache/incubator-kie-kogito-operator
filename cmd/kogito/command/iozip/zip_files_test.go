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

package iozip

import (
	"archive/tar"
	"compress/gzip"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/kiegroup/kogito-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-operator/cmd/kogito/command/test"
	"github.com/stretchr/testify/assert"
)

var baseTempDir = "/tmp/kogito-test/"

func writeFiles(t *testing.T, dirName string, files []string, content []byte) {
	for _, file := range files {
		err := ioutil.WriteFile(dirName+file, content, 0755)
		assert.Nil(t, err, "Error while writing the file")
		fileInfo, err := os.Stat(dirName + file)
		// double check file exists
		assert.Nil(t, err)
		assert.True(t, fileInfo.Size() > 0)
	}
}

func getFilesFromGzip(t *testing.T, fileToWrite *os.File) []string {
	// extract file and see if it is tarball file
	testTar, err := os.Open(fileToWrite.Name())
	assert.Nil(t, err)

	gzf, err := gzip.NewReader(testTar)
	assert.Nil(t, err)

	tarHeader := tar.NewReader(gzf)

	var filesFromGzip []string
	for {
		header, err := tarHeader.Next()
		if err == io.EOF {
			break
		}
		assert.Equal(t, header.Format, tar.FormatPAX)
		assert.NotNil(t, header.Name)
		filesFromGzip = append(filesFromGzip, header.Name)
	}
	return filesFromGzip
}

func TestCompressAsTGZS2i(t *testing.T) {
	tempDir := "/tmp/kogito-test/"
	simpleContent := []byte("hello World!!")
	files := []string{"file.bpmn", "file2.drl", "file3.bpmn2", "file4.dmn", "file5.properties", "file6.unsupported"}
	test.Mkdir(tempDir)
	writeFiles(t, tempDir, files, simpleContent)

	ioR, err := CompressAsTGZ(tempDir, flag.SourceToImageBuild)
	assert.Nil(t, err)
	fileToWrite, err1 := ioutil.TempFile(tempDir, "compressed_kogito_resources_*.tgz")
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

	filesFromGzip := getFilesFromGzip(t, fileToWrite)

	// all files extracted from gzip must be on the initial slice of test files
	for _, fileFromGzip := range filesFromGzip {
		assert.True(t, contains(fileFromGzip, files))
	}

	// And filesFromGzip should not contain the unsupported file: file6.unsupported
	assert.False(t, contains("file6.unsupported", filesFromGzip))

	os.RemoveAll(tempDir)
	os.Remove(fileToWrite.Name())
}

func TestCompressAsTGZQuarkusNative(t *testing.T) {
	libTempDir := baseTempDir + "lib/"
	test.Mkdir(libTempDir)

	simpleContent := []byte("hello World!!")
	baseFiles := []string{"file.json", "file2-runner", "file3.unsupported"}
	libFiles := []string{"file4.jar", "file5.unsupported"}

	writeFiles(t, baseTempDir, baseFiles, simpleContent)
	writeFiles(t, libTempDir, libFiles, simpleContent)

	ioR, err := CompressAsTGZ(baseTempDir, flag.BinaryQuarkusNativeBuild)
	assert.Nil(t, err)
	fileToWrite, err1 := ioutil.TempFile(baseTempDir, "compressed_kogito_resources_*.tgz")
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

	filesFromGzip := getFilesFromGzip(t, fileToWrite)
	filesInGzip := []string{"file.json", "file2-runner"}
	filesNotInGzip := []string{"file3.unsupported", "lib/file4.jar", "lib/file5.unsupported"}
	for _, file := range filesInGzip {
		assert.True(t, contains(file, filesFromGzip))
	}
	for _, file := range filesNotInGzip {
		assert.False(t, contains(file, filesFromGzip))
	}

	os.RemoveAll(baseTempDir)
}

func TestCompressAsTGZQuarkusFastJvm(t *testing.T) {
	libTempDir := baseTempDir + "lib/"
	quarkusFastJarTempDir := baseTempDir + "quarkus-app/"
	quarkusFastJarLibTempDir := quarkusFastJarTempDir + "lib/"
	quarkusFastJarLibMainTempDir := quarkusFastJarLibTempDir + "main/"
	quarkusFastJarLibBootTempDir := quarkusFastJarLibTempDir + "boot/"
	quarkusFastJarQuarkusTempDir := quarkusFastJarTempDir + "quarkus/"
	quarkusFastJarAppTempDir := quarkusFastJarTempDir + "app/"
	test.Mkdir(libTempDir)
	test.Mkdir(quarkusFastJarTempDir)
	test.Mkdir(quarkusFastJarLibTempDir)
	test.Mkdir(quarkusFastJarLibMainTempDir)
	test.Mkdir(quarkusFastJarLibBootTempDir)
	test.Mkdir(quarkusFastJarQuarkusTempDir)
	test.Mkdir(quarkusFastJarAppTempDir)

	simpleContent := []byte("hello World!!")
	baseFiles := []string{"file.json", "file2-runner", "file3-runner.jar", "file4.unsupported"}
	libFiles := []string{"file5.jar", "file6.unsupported"}
	quarkusFastJarLibMainFiles := []string{"filelibmain.jar", "filelibmain.unsupported"}
	quarkusFastJarLibBootFiles := []string{"filelibboot.jar", "filelibboot.unsupported"}
	quarkusFastJarQuarkusFiles := []string{"file9.dat"}
	quarkusFastJarFiles := []string{"quarkus-run.jar", "file11.unsupported"}

	writeFiles(t, baseTempDir, baseFiles, simpleContent)
	writeFiles(t, libTempDir, libFiles, simpleContent)
	writeFiles(t, quarkusFastJarLibMainTempDir, quarkusFastJarLibMainFiles, simpleContent)
	writeFiles(t, quarkusFastJarLibBootTempDir, quarkusFastJarLibBootFiles, simpleContent)
	writeFiles(t, quarkusFastJarQuarkusTempDir, quarkusFastJarQuarkusFiles, simpleContent)
	writeFiles(t, quarkusFastJarTempDir, quarkusFastJarFiles, simpleContent)

	// ensure files in nested dirs work without trailing slash
	ioR, err := CompressAsTGZ(strings.TrimSuffix(baseTempDir, "/"), flag.BinaryQuarkusFastJarJvmBuild)
	assert.Nil(t, err)
	fileToWrite, err1 := ioutil.TempFile(baseTempDir, "compressed_kogito_resources_*.tgz")
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

	filesFromGzip := getFilesFromGzip(t, fileToWrite)
	filesInGzip := []string{"file.json", "file3-runner.jar", "quarkus-app/quarkus-run.jar", "quarkus-app/lib/main/filelibmain.jar", "quarkus-app/lib/boot/filelibboot.jar", "quarkus-app/quarkus/file9.dat"}
	filesNotInGzip := []string{"file2-runner", "file4.unsupported", "lib/file6.unsupported", "lib/file5.jar", "quarkus-app/file11.unsupported", "quarkus-app/lib/main/filelibmain.unsupported", "quarkus-app/lib/boot/filelibboot.unsupported"}
	for _, file := range filesInGzip {
		assert.True(t, contains(file, filesFromGzip))
	}
	for _, file := range filesNotInGzip {
		assert.False(t, contains(file, filesFromGzip))
	}

	os.RemoveAll(baseTempDir)
}

func TestCompressAsTGZQuarkusLegacyJvm(t *testing.T) {
	libTempDir := baseTempDir + "lib/"
	quarkusLibTempDir := baseTempDir + "quarkus-app/lib/"
	test.Mkdir(libTempDir)
	test.Mkdir(quarkusLibTempDir)

	simpleContent := []byte("hello World!!")
	baseFiles := []string{"file.json", "file2-runner", "file3-runner.jar", "file4.unsupported"}
	libFiles := []string{"file5.jar", "file6.unsupported"}
	quarkusLibFiles := []string{"file7.jar", "file8.unsupported"}

	writeFiles(t, baseTempDir, baseFiles, simpleContent)
	writeFiles(t, libTempDir, libFiles, simpleContent)
	writeFiles(t, quarkusLibTempDir, quarkusLibFiles, simpleContent)

	// ensure files in nested dirs work without trailing slash
	ioR, err := CompressAsTGZ(strings.TrimSuffix(baseTempDir, "/"), flag.BinaryQuarkusLegacyJarJvmBuild)
	assert.Nil(t, err)
	fileToWrite, err1 := ioutil.TempFile(baseTempDir, "compressed_kogito_resources_*.tgz")
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

	filesFromGzip := getFilesFromGzip(t, fileToWrite)
	filesInGzip := []string{"file.json", "file3-runner.jar", "lib/file5.jar"}
	filesNotInGzip := []string{"file2-runner", "file4.unsupported", "lib/file6.unsupported", "quarkus-app/lib/file7.jar", "quarkus-app/lib/file8.unsupported"}
	for _, file := range filesInGzip {
		assert.True(t, contains(file, filesFromGzip))
	}
	for _, file := range filesNotInGzip {
		assert.False(t, contains(file, filesFromGzip))
	}

	os.RemoveAll(baseTempDir)
}

func TestCompressAsTGZSpringBootJvm(t *testing.T) {
	libTempDir := baseTempDir + "lib/"
	quarkusLibTempDir := baseTempDir + "quarkus-app/lib/"
	test.Mkdir(libTempDir)
	test.Mkdir(quarkusLibTempDir)

	simpleContent := []byte("hello World!!")
	baseFiles := []string{"file.json", "file2-runner", "file3.jar", "file4.unsupported"}
	libFiles := []string{"file5.jar", "file6.unsupported"}
	quarkusLibFiles := []string{"file7.jar", "file8.unsupported"}

	writeFiles(t, baseTempDir, baseFiles, simpleContent)
	writeFiles(t, libTempDir, libFiles, simpleContent)
	writeFiles(t, quarkusLibTempDir, quarkusLibFiles, simpleContent)

	ioR, err := CompressAsTGZ(baseTempDir, flag.BinarySpringBootJvmBuild)
	assert.Nil(t, err)
	fileToWrite, err1 := ioutil.TempFile(baseTempDir, "compressed_kogito_resources_*.tgz")
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

	filesFromGzip := getFilesFromGzip(t, fileToWrite)
	filesInGzip := []string{"file.json", "file3.jar"}
	filesNotInGzip := []string{"file2-runner", "file4.unsupported", "lib/file5.jar", "lib/file6.unsupported", "quarkus-app/lib/file7.jar", "quarkus-app/lib/file8.unsupported"}
	for _, file := range filesInGzip {
		assert.True(t, contains(file, filesFromGzip))
	}
	for _, file := range filesNotInGzip {
		assert.False(t, contains(file, filesFromGzip))
	}

	os.RemoveAll(baseTempDir)
}

func TestIsSuffixSupported(t *testing.T) {
	//test a few invalid extensions
	invalidExtensions := []string{"test.dn", "test.st", "test.www", "test.bmn2", "test.prop"}
	for _, ext := range invalidExtensions {
		assert.False(t, IsSuffixSupported(ext, flag.SourceToImageBuild))
	}

	//test a few valid extensions
	validExtensions := []string{"test.drl", "test.dmn", "test.bpmn2", "test.properties", "test.bpmn"}
	for _, ext := range validExtensions {
		assert.True(t, IsSuffixSupported(ext, flag.SourceToImageBuild))
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
