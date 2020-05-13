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

import "fmt"

type KogitoChannelType string

const (
	// alpha channel
	AlphaChannel KogitoChannelType = "alpha"

	// dev-preview channel
	DevPreviewChannel KogitoChannelType = "dev-preview"
)

func (s *KogitoChannelType) Set(val string) error {
	*s = KogitoChannelType(val)
	return nil
}
func (s *KogitoChannelType) Type() string {
	return "string"
}

func (s *KogitoChannelType) String() string { return string(*s) }

func (ch KogitoChannelType) IsValid() error {
	switch ch {
	case AlphaChannel, DevPreviewChannel:
		return nil
	}
	return fmt.Errorf("Invalid Kogito channel type : %s ", ch)
}
