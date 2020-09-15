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

package converter

import (
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_FromImageTagToImage(t *testing.T) {
	buildImage := "mydomain.io/mynamespace/builder-image:1.0"
	image := FromImageTagToImage(buildImage, "2.0")
	assert.NotNil(t, image)
	assert.Equal(t, "mydomain.io", image.Domain)
	assert.Equal(t, "mynamespace", image.Namespace)
	assert.Equal(t, "builder-image", image.Name)
	assert.Equal(t, "1.0", image.Tag)
}

func Test_FromImageTagToImage_EmptyImageTag(t *testing.T) {
	image := FromImageTagToImage("", "2.0")
	assert.NotNil(t, image)
	assert.Equal(t, "", image.Domain)
	assert.Equal(t, "", image.Namespace)
	assert.Equal(t, "", image.Name)
	assert.Equal(t, "2.0", image.Tag)
}

func Test_FromImageFlagToImage(t *testing.T) {
	imageFlags := flag.ImageFlags{
		Image:                 "mydomain.io/mynamespace/builder-image:1.0",
		InsecureImageRegistry: true,
	}

	image := FromImageFlagToImage(&imageFlags)
	assert.NotNil(t, image)
	assert.Equal(t, "mydomain.io", image.Domain)
	assert.Equal(t, "mynamespace", image.Namespace)
	assert.Equal(t, "builder-image", image.Name)
	assert.Equal(t, "1.0", image.Tag)
}

func Test_FromImageTagWithPortToImage(t *testing.T) {
	buildImage := "mydomain.io:5050/mynamespace/builder-image:1.0"
	image := FromImageTagToImage(buildImage, "2.0")
	assert.NotNil(t, image)
	assert.Equal(t, "mydomain.io:5050", image.Domain)
	assert.Equal(t, "mynamespace", image.Namespace)
	assert.Equal(t, "builder-image", image.Name)
	assert.Equal(t, "1.0", image.Tag)
}

func Test_FromImageTagWithPortNoTagToImage(t *testing.T) {
	buildImage := "mydomain.io:5050/mynamespace/builder-image"
	image := FromImageTagToImage(buildImage, "2.0")
	assert.NotNil(t, image)
	assert.Equal(t, "mydomain.io:5050", image.Domain)
	assert.Equal(t, "mynamespace", image.Namespace)
	assert.Equal(t, "builder-image", image.Name)
	assert.Equal(t, "latest", image.Tag)
}

func Test_FromImageTagWithPortNoLocalhostTagToImage(t *testing.T) {
	buildImage := "localhost:5050/mynamespace/builder-image"
	image := FromImageTagToImage(buildImage, "2.0")
	assert.NotNil(t, image)
	assert.Equal(t, "localhost:5050", image.Domain)
	assert.Equal(t, "mynamespace", image.Namespace)
	assert.Equal(t, "builder-image", image.Name)
	assert.Equal(t, "latest", image.Tag)
}
