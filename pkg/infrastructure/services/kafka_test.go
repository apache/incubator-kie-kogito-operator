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

package services

import "testing"

func Test_fromKafkaTopicToQuarkusEnvVar(t *testing.T) {
	type args struct {
		topic KafkaTopicDefinition
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Incoming Topic", args{topic: KafkaTopicDefinition{TopicName: "kogito-processinstances-events", MessagingType: KafkaTopicIncoming}}, "mp.messaging.incoming.kogito-processinstances-events.bootstrap.servers"},
		{"Outgoing Topic", args{topic: KafkaTopicDefinition{TopicName: "kogito-job-service-job-status-events", MessagingType: KafkaTopicOutgoing}}, "mp.messaging.outgoing.kogito-job-service-job-status-events.bootstrap.servers"},
		{"Blank", args{topic: KafkaTopicDefinition{TopicName: "", MessagingType: ""}}, ""},
		{"Nil", args{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fromKafkaTopicToQuarkusAppProp(tt.args.topic); got != tt.want {
				t.Errorf("fromKafkaTopicToQuarkusAppProp() = %v, want %v", got, tt.want)
			}
		})
	}
}
