// Copyright 2019 Red Hat, Inc. and/or its affiliates
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

package framework

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/kiegroup/kogito-cloud-operator/test/config"
)

// GenerateNamespaceName generates a namespace name, taking configuration into account (local or not)
func GenerateNamespaceName(prefix string) string {
	rand.Seed(time.Now().UnixNano())
	ns := fmt.Sprintf("%s-%s", prefix, test.GenerateShortUID(4))
	if config.IsLocalTests() {
		username := getEnvUsername()
		ns = fmt.Sprintf("%s-local-%s", username, ns)
	} else if len(config.GetCiName()) > 0 {
		ns = fmt.Sprintf("%s-%s", config.GetCiName(), ns)
	}
	return ns
}

// ReadFromURI reads string content from given URI (URL or Filesystem)
func ReadFromURI(uri string) (string, error) {
	var data []byte
	if strings.HasPrefix(uri, "http") {
		resp, err := http.Get(uri)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if data, err = ioutil.ReadAll(resp.Body); err != nil {
			return "", err
		}
	} else {
		// It should be a Filesystem uri
		absPath, err := filepath.Abs(uri)
		if err != nil {
			return "", err
		}
		data, err = ioutil.ReadFile(absPath)
		if err != nil {
			return "", err
		}
	}
	return string(data), nil
}

// WaitFor waits for a specification condition to be met or until one error condition is met
func WaitFor(namespace, display string, timeout time.Duration, condition func() (bool, error), errorConditions ...func() (bool, error)) error {
	GetLogger(namespace).Info(fmt.Sprintf("Wait %s for %s", timeout.String(), display))

	timeoutChan := time.After(timeout)
	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("Timeout waiting for %s", display)
		case <-tick.C:
			running, err := condition()
			if err != nil {
				GetLogger(namespace).Warn(fmt.Sprintf("Problem in condition execution, waiting for %s => %v", display, err))
			}

			if running {
				GetLogger(namespace).Info(fmt.Sprintf("'%s' is successful", display))
				return nil
			}

			for _, errorCondition := range errorConditions {
				if hasErrors, err := errorCondition(); hasErrors {
					GetLogger(namespace).Error(err, "Problem in condition execution", "display", display)
					return err
				}
			}
		}
	}
}

// PrintDataMap prints a formatted dataMap using the given writer
func PrintDataMap(keys []string, dataMaps []map[string]string, writer io.StringWriter) error {
	// Get size of strings to be written, to be able to format correctly
	maxStringSizeMap := make(map[string]int)
	for _, key := range keys {
		maxSize := len(key)
		for _, dataMap := range dataMaps {
			if len(dataMap[key]) > maxSize {
				maxSize = len(dataMap[key])
			}
		}
		maxStringSizeMap[key] = maxSize
	}

	// Write headers
	for _, header := range keys {
		if _, err := writer.WriteString(header); err != nil {
			return fmt.Errorf("Error in writing the header: %v", err)
		}
		if _, err := writer.WriteString(getWhitespaceStr(maxStringSizeMap[header] - len(header) + 1)); err != nil {
			return fmt.Errorf("Error in writing headers: %v", err)
		}
		if _, err := writer.WriteString(" | "); err != nil {
			return fmt.Errorf("Error in writing headers : %v", err)
		}
	}
	if _, err := writer.WriteString("\n"); err != nil {
		return fmt.Errorf("Error in writing headers '|': %v", err)

	}

	// Write events
	for _, dataMap := range dataMaps {
		for _, key := range keys {
			if _, err := writer.WriteString(dataMap[key]); err != nil {
				return fmt.Errorf("Error in writing events: %v", err)
			}
			if _, err := writer.WriteString(getWhitespaceStr(maxStringSizeMap[key] - len(dataMap[key]) + 1)); err != nil {
				return fmt.Errorf("Error in writing events: %v", err)
			}
			if _, err := writer.WriteString(" | "); err != nil {
				return fmt.Errorf("Error in writing events: %v", err)
			}
		}
		if _, err := writer.WriteString("\n"); err != nil {
			return fmt.Errorf("Error in writing events: %v", err)
		}
	}
	return nil
}

func getWhitespaceStr(size int) string {
	whiteSpaceStr := ""
	for i := 0; i < size; i++ {
		whiteSpaceStr += " "
	}
	return whiteSpaceStr
}

// CreateFolder  creates a folder and all its parents if not exist
func CreateFolder(folder string) error {
	return os.MkdirAll(folder, os.ModePerm)
}

// CreateTemporaryFolder creates a folder in default directory for temporary files
func CreateTemporaryFolder(folderPrefix string) (string, error) {
	return ioutil.TempDir("", folderPrefix)
}

// DeleteFolder deletes a folder and all its subfolders
func DeleteFolder(folder string) error {
	return os.RemoveAll(folder)
}

// CreateFile Creates file in folder with supplied content
func CreateFile(folder, fileName, fileContent string) error {
	f, err := os.Create(folder + "/" + fileName)
	if err != nil {
		return fmt.Errorf("Error creating file %s in folder %s: %v ", fileName, folder, err)
	}

	if _, err = f.WriteString(fileContent); err != nil {
		f.Close()
		return fmt.Errorf("Error writing to file %s in folder %s: %v ", fileName, folder, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("Error closing file %s in folder %s: %v ", fileName, folder, err)
	}
	return nil
}

// CreateTemporaryFile Creates file in default directory for temporary files with supplied content
func CreateTemporaryFile(filePattern, fileContent string) (string, error) {
	f, err := ioutil.TempFile("", filePattern)
	if err != nil {
		return "", fmt.Errorf("Error creating file with pattern %s in temporary folder: %v ", filePattern, err)
	}

	if _, err = f.WriteString(fileContent); err != nil {
		f.Close()
		return "", fmt.Errorf("Error writing to file %s in temporary folder: %v ", f.Name(), err)
	}

	if err := f.Close(); err != nil {
		return "", fmt.Errorf("Error closing file %s in temporary folder: %v ", f.Name(), err)
	}

	return f.Name(), nil
}

// DeleteFile deletes a file
func DeleteFile(folder, fileName string) error {
	return os.Remove(folder + "/" + fileName)
}

// GetBuildImage returns a build image with defaults set
func GetBuildImage(imageName string) string {
	image := v1beta1.Image{
		Domain:    config.GetBuildImageRegistry(),
		Namespace: config.GetBuildImageNamespace(),
		Name:      imageName,
		Tag:       config.GetBuildImageVersion(),
	}

	if len(image.Domain) == 0 {
		image.Domain = infrastructure.DefaultImageRegistry
	}

	if len(image.Namespace) == 0 {
		image.Namespace = infrastructure.DefaultImageNamespace
	}

	if len(image.Tag) == 0 {
		image.Tag = infrastructure.GetKogitoImageVersion()
	}

	// Update image name with suffix if provided
	if len(config.GetBuildImageNameSuffix()) > 0 {
		image.Name = fmt.Sprintf("%s-%s", image.Name, config.GetBuildImageNameSuffix())
	}

	return framework.ConvertImageToImageTag(image)
}
