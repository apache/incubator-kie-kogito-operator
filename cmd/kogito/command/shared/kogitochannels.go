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

// KogitoChannelType describes the supported OLM channels that the Kogito Operator can be installed
type KogitoChannelType string

const (
	// AlphaChannel ...
	AlphaChannel KogitoChannelType = "alpha"

	// DevPreviewChannel ...
	DevPreviewChannel KogitoChannelType = "dev-preview"
)

// IsChannelValid validates user provide channel value is among valid channel. If channel is value then return true else false
func IsChannelValid(ch string) bool {
	inputChannel := KogitoChannelType(ch)
	switch inputChannel {
	case AlphaChannel, DevPreviewChannel:
		return true
	}
	return false
}

// GetDefaultChannel provides default chanel
func GetDefaultChannel() KogitoChannelType {
	return AlphaChannel
}
