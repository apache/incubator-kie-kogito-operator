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

package shared

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/util"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// ResourceType represents mediums through which user can trigger build
type ResourceType string

const (
	// LocalFileResource Build using file on local system
	LocalFileResource ResourceType = "LocalFileResource"
	// LocalDirectoryResource Build using directory on local system
	LocalDirectoryResource ResourceType = "LocalDirectoryResource"
	// GitFileResource Build using file in Git Repo
	GitFileResource ResourceType = "GitFileResource"
	// GitRepositoryResource Build using Git Repo
	GitRepositoryResource ResourceType = "GitRepositoryResource"
	// BinaryResource Build using user generated binaries
	BinaryResource ResourceType = "BinaryResource"
)

// GetResourceType drive resource type using provide resource URI.
// If resource URI is not provided then its a Binary build request
// If resource URI starts with HTTP and end with file ext suffix then its a build request using Git file
// If resource URI starts with HTTP and don't have file ext suffix then its a build request using Git Repo
// If resource URI is refers to local system and end with file ext suffix then its a build request using local file
// If resource URI is refers to local system and don't file ext suffix then its a build request using local directory
func GetResourceType(resource string) (ResourceType ResourceType, err error) {

	// check for binary resource
	if len(resource) == 0 {
		return BinaryResource, nil
	}

	// check for Git resource
	if strings.HasPrefix(resource, "http") {
		parsedURL, err := url.ParseRequestURI(resource)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			return "", &url.Error{URL: resource, Err: err}
		}
		// check whether resource is Git File or Git Repo
		ff := strings.Split(resource, "/")
		fileName := strings.Join(strings.Fields(ff[len(ff)-1]), "")
		if util.IsSuffixSupported(fileName) {
			return GitFileResource, nil
		}
		return GitRepositoryResource, nil
	}

	// check for local resource
	fileInfo, err := os.Stat(resource)
	if err != nil {
		return
	}
	if fileInfo.Mode().IsRegular() {
		if util.IsSuffixSupported(resource) {
			return LocalFileResource, nil
		}
		return "", fmt.Errorf("invalid resource")
	} else if fileInfo.Mode().IsDir() {
		return LocalDirectoryResource, nil
	}

	return "", fmt.Errorf("invalid resource")
}

// LoadGitFileIntoMemory reads file from remote Git location and load it in memory.
func LoadGitFileIntoMemory(resource string) (io.Reader, string, error) {
	log := context.GetDefaultLogger()
	parsedURL, _ := url.ParseRequestURI(resource)
	ff := strings.Split(parsedURL.Path, "/")
	fileName := strings.Join(strings.Fields(ff[len(ff)-1]), "")

	response, err := http.Get(resource)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download %s, error message: %s", resource, err.Error())
	}
	log.Infof(message.KogitoBuildFoundAsset, fileName)
	return response.Body, fileName, nil
}

// LoadLocalFileIntoMemory reads file from local system and load it in memory.
func LoadLocalFileIntoMemory(resource string) (io.Reader, string, error) {
	log := context.GetDefaultLogger()
	log.Infof(message.KogitoBuildFoundFile, resource)
	ff := strings.Split(resource, "/")
	fileName := strings.Join(strings.Fields(ff[len(ff)-1]), "")
	fileReader, err := os.Open(resource)
	if err != nil {
		return nil, "", err
	}
	return fileReader, fileName, nil
}

// ZipAndLoadLocalDirectoryIntoMemory zip the given directory URI and load it in memory.
func ZipAndLoadLocalDirectoryIntoMemory(resource string) (io.Reader, string, error) {
	log := context.GetDefaultLogger()
	log.Info(message.KogitoBuildProvidedFileIsDir)
	ioTgzR, err := util.ProduceTGZfile(resource)
	if err != nil {
		return nil, "", err
	}
	return ioTgzR, "", nil
}
