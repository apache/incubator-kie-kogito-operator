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
	"bytes"
	"compress/gzip"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var supportedExtensions = []string{".dmn", ".drl", ".bpmn", ".bpmn2", ".properties"}

// ProduceTGZfile produces a tgz file of the given directory files
func ProduceTGZfile(resource string) (io.Reader, error) {
	log := context.GetDefaultLogger()
	var buf bytes.Buffer
	var filesFound []string
	var link string

	gzipWriter := gzip.NewWriter(&buf)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	err := filepath.Walk(resource, func(absoluteFilePath string, fileInfo os.FileInfo, err error) error {
		if IsSuffixSupported(fileInfo.Name()) {

			fileToCompress, err := os.Open(absoluteFilePath)
			if err != nil {
				return err
			}
			defer fileToCompress.Close()

			stat, err := fileToCompress.Stat()
			if err != nil {
				return err
			}

			if fileInfo.Mode()&os.ModeSymlink != 0 {
				link, err = os.Readlink(absoluteFilePath)
				if err != nil {
					return err
				}
			}
			header, err := tar.FileInfoHeader(fileInfo, link)
			if err != nil {
				return err
			}
			header.Name = filepath.ToSlash(fileInfo.Name())
			header.Linkname = filepath.ToSlash(header.Linkname)
			header.Format = tar.FormatPAX
			header.Size = stat.Size()
			header.Mode = int64(stat.Mode())
			header.ModTime = stat.ModTime()
			// write header
			if err := tarWriter.WriteHeader(header); err != nil {
				return err
			}

			if _, err := io.Copy(tarWriter, fileToCompress); err != nil {
				return err
			}

			filesFound = append(filesFound, fileToCompress.Name())
		}
		return nil
	})

	if err != nil {
		log.Errorf(message.KogitoBuildFileWalkingError, resource, err)
	}
	log.Infof(message.KogitoBuildFoundFile, filesFound)
	return &buf, err
}

// IsSuffixSupported checks if the given extension is supported
// when performing a build from single file
func IsSuffixSupported(value string) bool {
	for _, ext := range supportedExtensions {
		if strings.HasSuffix(value, ext) {
			return true
		}
	}
	return false
}
