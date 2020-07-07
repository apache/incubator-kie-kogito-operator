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

package flag

import (
	buildutil "github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/util"
	"github.com/spf13/cobra"
)

// ImageFlags is common properties used to configure Image
type ImageFlags struct {
	Image                 string
	InsecureImageRegistry bool
}

// AddImageFlags adds the image flags to the given command
func AddImageFlags(command *cobra.Command, flags *ImageFlags) {
	command.Flags().StringVarP(&flags.Image, "image", "i", "", "The image which should be used to run Service. For example 'quay.io/kiegroup/kogito-data-index:latest'")
	command.Flags().BoolVar(&flags.InsecureImageRegistry, "insecure-image-registry", false, "Indicates that the Service image points to insecure image registry")
}

// CheckImageArgs validates the ImageFlags flags
func CheckImageArgs(flags *ImageFlags) error {
	if err := buildutil.CheckImageTag(flags.Image); err != nil {
		return err
	}
	return nil
}

// IsEmpty return true if image details are not provided else return false
func (i *ImageFlags) IsEmpty() bool {
	return len(i.Image) == 0
}
