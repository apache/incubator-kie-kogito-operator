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

package deploy

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_DeleteServiceCmd_SuccessfullyDelete(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("delete-service example-drools --project %s", ns)
	test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&v1beta1.KogitoRuntime{ObjectMeta: metav1.ObjectMeta{Name: "example-drools", Namespace: ns}})

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "Successfully deleted Kogito Service example-drools")
}

func Test_DeleteServiceCmd_Failure_ServiceDoesNotExist(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("delete-service example-drools --project %s", ns)
	test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	_, errLines, err := test.ExecuteCli()
	assert.Error(t, err)
	assert.Contains(t, errLines, "with the name 'example-drools' doesn't exist")
}
