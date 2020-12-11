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
	"bytes"
	"compress/gzip"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var supportedExtensions = map[flag.BinaryBuildType][]string{
	flag.SourceToImageBuild:       {".dmn", ".drl", ".bpmn", ".bpmn2", ".properties", ".sw.json", ".sw.yaml"},
	flag.BinaryQuarkusJvmBuild:    {".jar", ".json"},
	flag.BinarySpringBootJvmBuild: {".jar", ".json"},
	flag.BinaryQuarkusNativeBuild: {"-runner", ".json"},
}

func zipFile(absoluteFilePath string, fileInfo os.FileInfo, resource string, binaryBuildType flag.BinaryBuildType, tarWriter *tar.Writer) (string, error) {
	var link string

	fileToCompress, err := os.Open(absoluteFilePath)
	if err != nil {
		return "", err
	}
	defer fileToCompress.Close()

	stat, err := fileToCompress.Stat()
	if err != nil {
		return "", err
	}

	if fileInfo.Mode()&os.ModeSymlink != 0 {
		link, err = os.Readlink(absoluteFilePath)
		if err != nil {
			return "", err
		}
	}
	header, err := tar.FileInfoHeader(fileInfo, link)
	if err != nil {
		return "", err
	}

	if binaryBuildType == flag.SourceToImageBuild {
		// don't include directory structure if not binary build
		header.Name = filepath.ToSlash(fileInfo.Name())
	} else {
		// include directory structure inside base dir if binary build (don't include base dir)
		// since s2i script looks for lib/ dir
		header.Name = strings.TrimPrefix(absoluteFilePath, resource)
	}
	header.Linkname = filepath.ToSlash(header.Linkname)
	header.Format = tar.FormatPAX
	header.Size = stat.Size()
	header.Mode = int64(stat.Mode())
	header.ModTime = stat.ModTime()
	// write header
	if err := tarWriter.WriteHeader(header); err != nil {
		return "", err
	}

	if _, err := io.Copy(tarWriter, fileToCompress); err != nil {
		return "", err
	}

	return fileToCompress.Name(), nil
}

func zipFilesInDir(dir string, resource string, binaryBuildType flag.BinaryBuildType, tarWriter *tar.Writer) ([]string, error) {
	var filesFound []string
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, fileInfo := range files {
		if IsSuffixSupported(fileInfo.Name(), binaryBuildType) && !fileInfo.IsDir() {
			zippedFileName, err := zipFile(dir+fileInfo.Name(), fileInfo, resource, binaryBuildType, tarWriter)
			if err != nil {
				return filesFound, err
			}
			filesFound = append(filesFound, zippedFileName)
		}
	}
	return filesFound, nil
}

// CompressAsTGZ produces a tgz file of the given directory files
func CompressAsTGZ(resource string, binaryBuildType flag.BinaryBuildType) (io.Reader, error) {
	log := context.GetDefaultLogger()
	var buf bytes.Buffer
	var filesFound []string
	if !strings.HasSuffix(resource, "/") {
		resource += "/"
	}
	var err error

	gzipWriter := gzip.NewWriter(&buf)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// traverse all nested directories for Kogito resources for s2i builds
	if binaryBuildType == flag.SourceToImageBuild {
		err = filepath.Walk(resource, func(absoluteFilePath string, fileInfo os.FileInfo, walkErr error) error {
			if IsSuffixSupported(fileInfo.Name(), binaryBuildType) && !fileInfo.IsDir() {
				zippedFileName, err := zipFile(absoluteFilePath, fileInfo, resource, binaryBuildType, tarWriter)
				if err != nil {
					return err
				}
				filesFound = append(filesFound, zippedFileName)
			}
			return nil
		})
		// only look in base directory for supported extensions for other builds
	} else {
		filesFound, err = zipFilesInDir(resource, resource, binaryBuildType, tarWriter)
	}
	// look in lib dir as well for Quarkus JVM builds
	if binaryBuildType == flag.BinaryQuarkusJvmBuild {
		var libFilesFound []string
		libFilesFound, err = zipFilesInDir(resource+"lib/", resource, binaryBuildType, tarWriter)
		filesFound = append(filesFound, libFilesFound...)
	}

	if err != nil {
		log.Errorf(message.KogitoBuildFileWalkingError, resource, err)
	}
	log.Infof(message.KogitoBuildFoundFile, filesFound)
	return &buf, err
}

// IsSuffixSupported checks if the given extension is supported
// when performing a build from single file
func IsSuffixSupported(value string, binaryBuildType flag.BinaryBuildType) bool {
	for _, ext := range supportedExtensions[binaryBuildType] {
		if strings.HasSuffix(value, ext) {
			return true
		}
	}
	return false
}
