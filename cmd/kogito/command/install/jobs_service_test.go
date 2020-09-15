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

package install

import (
	"fmt"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_DeployJobServiceCmd_DefaultConfiguration(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install jobs-service --project %s", ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	lines, _, err := test.ExecuteCli()

	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Jobs Service successfully installed")

	// This should be created, given the command above
	jobService := &v1alpha1.KogitoJobsService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.DefaultJobsServiceName,
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(ctx.Client).Fetch(jobService)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.NotNil(t, jobService)
	assert.False(t, jobService.Spec.InsecureImageRegistry)
	assert.Equal(t, int64(0), jobService.Spec.BackOffRetryMillis)
	assert.Equal(t, int64(0), jobService.Spec.MaxIntervalLimitToRetryMillis)
}

func Test_DeployJobServiceCmd_CustomConfiguration(t *testing.T) {
	ns := t.Name()
	cli := fmt.Sprintf("install jobs-service --project %s --enable-events --enable-persistence --infinispan-url myservice:11222 --kafka-url my-cluster:9092 --infinispan-user user --infinispan-password password --insecure-image-registry --http-port 9090 --backoff-retry-millis=5 --max-internal-limit-retry-millis=10", ns)
	ctx := test.SetupCliTest(cli,
		context.CommandFactory{BuildCommands: BuildCommands},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	lines, _, err := test.ExecuteCli()

	assert.NoError(t, err)
	assert.Contains(t, lines, "Kogito Jobs Service successfully installed")

	// This should be created, given the command above
	jobService := &v1alpha1.KogitoJobsService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      infrastructure.DefaultJobsServiceName,
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(ctx.Client).Fetch(jobService)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.NotNil(t, jobService)
	assert.True(t, jobService.Spec.InsecureImageRegistry)
	assert.Equal(t, int32(9090), jobService.Spec.HTTPPort)
	assert.Equal(t, int64(5), jobService.Spec.BackOffRetryMillis)
	assert.Equal(t, int64(10), jobService.Spec.MaxIntervalLimitToRetryMillis)
}
