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

package remove

import (
	"fmt"
	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/api/v1beta1"
	"github.com/kiegroup/kogito-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-operator/cmd/kogito/command/test"
	"github.com/kiegroup/kogito-operator/core/kogitosupportingservice"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_removeRuntimeServiceCommand_NoServiceThere(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("remove data-index -p %s", ns)
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "There's no service data-index")
}

func Test_removeRuntimeServiceCommand_NoServiceThereWithAlias(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("remove management-console -p %s", ns)
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, "There's no service mgmt-console")
}

func Test_removeRuntimeServiceCommand_SingletonService(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("remove jobs-service -p %s", ns)
	jobsService := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{Name: kogitosupportingservice.DefaultJobsServiceName, Namespace: ns},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: api.JobsService,
		},
	}
	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}, jobsService)

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.NotContains(t, lines, "There's no service jobs-service")
	assert.Contains(t, lines, kogitosupportingservice.DefaultJobsServiceName)
	assert.Contains(t, lines, "has been successfully removed")
}

func Test_removeRuntimeServiceCommand_MoreThenOneService(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("remove jobs-service -p %s", ns)
	jobsService1 := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{Name: kogitosupportingservice.DefaultJobsServiceName, Namespace: ns},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: api.JobsService,
		},
	}

	jobsService2 := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{Name: "my-job-service", Namespace: ns},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: api.JobsService,
		},
	}

	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}, jobsService1, jobsService2)

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, kogitosupportingservice.DefaultJobsServiceName)
	assert.Contains(t, lines, "my-job-service")
	assert.Contains(t, lines, "has been successfully removed")
}

func Test_removeRuntimeServiceCommand_DifferentService(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("remove jobs-service -p %s", ns)
	jobsService1 := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{Name: kogitosupportingservice.DefaultJobsServiceName, Namespace: ns},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: api.JobsService,
		},
	}

	jobsService2 := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{Name: "data-index", Namespace: ns},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: api.DataIndex,
		},
	}

	test.SetupCliTest(cli, context.CommandFactory{BuildCommands: BuildCommands}, &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}, jobsService1, jobsService2)

	lines, _, err := test.ExecuteCli()
	assert.NoError(t, err)
	assert.Contains(t, lines, kogitosupportingservice.DefaultJobsServiceName)
	assert.NotContains(t, lines, "data-index")
	assert.Contains(t, lines, "has been successfully removed")
}
