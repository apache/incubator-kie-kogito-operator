// Copyright 2021 Red Hat, Inc. and/or its affiliates
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

package api

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// KogitoBuildType describes the build types supported by the KogitoBuild CR
type KogitoBuildType string

const (
	// BinaryBuildType builds takes an uploaded binary file already compiled and creates a Kogito service image from it.
	BinaryBuildType KogitoBuildType = "Binary"
	// RemoteSourceBuildType builds pulls the source code from a Git repository, builds the binary and then the final Kogito service image.
	RemoteSourceBuildType KogitoBuildType = "RemoteSource"
	// LocalSourceBuildType builds takes an uploaded resource files such as DRL (rules), DMN (decision) or BPMN (process), builds the binary and the final Kogito service image.
	LocalSourceBuildType KogitoBuildType = "LocalSource"
)

// KogitoBuildConditionType ...
type KogitoBuildConditionType string

const (
	// KogitoBuildSuccessful condition for a successful build.
	KogitoBuildSuccessful KogitoBuildConditionType = "Successful"
	// KogitoBuildFailure condition for a failure build.
	KogitoBuildFailure KogitoBuildConditionType = "Failed"
	// KogitoBuildRunning condition for a running build.
	KogitoBuildRunning KogitoBuildConditionType = "Running"
)

// KogitoBuildConditionReason ...
type KogitoBuildConditionReason string

const (
	// OperatorFailureReason when operator fails to reconcile.
	OperatorFailureReason KogitoBuildConditionReason = "OperatorFailure"
	// BuildFailureReason when build fails.
	BuildFailureReason KogitoBuildConditionReason = "BuildFailure"
)

// KogitoBuildInterface ...
type KogitoBuildInterface interface {
	metav1.Object
	runtime.Object
	// GetSpec gets the Kogito Service specification structure.
	GetSpec() KogitoBuildSpecInterface
	// GetStatus gets the Kogito Service Status structure.
	GetStatus() KogitoBuildStatusInterface
}

// KogitoBuildSpecInterface ...
type KogitoBuildSpecInterface interface {
	GetType() KogitoBuildType
	SetType(buildType KogitoBuildType)
	IsDisableIncremental() bool
	SetDisableIncremental(disableIncremental bool)
	GetEnv() []corev1.EnvVar
	SetEnv(env []corev1.EnvVar)
	GetGitSource() GitSourceInterface
	SetGitSource(gitSource GitSourceInterface)
	GetRuntime() RuntimeType
	SetRuntime(runtime RuntimeType)
	GetWebHooks() []WebHookSecretInterface
	SetWebHooks(webhooks []WebHookSecretInterface)
	IsNative() bool
	SetNative(native bool)
	GetResources() corev1.ResourceRequirements
	SetResources(resources corev1.ResourceRequirements)
	GetMavenMirrorURL() string
	SetMavenMirrorURL(mavenMirrorURL string)
	GetBuildImage() string
	SetBuildImage(buildImage string)
	GetRuntimeImage() string
	SetRuntimeImage(runtime string)
	GetTargetKogitoRuntime() string
	SetTargetKogitoRuntime(targetRuntime string)
	GetArtifact() ArtifactInterface
	SetArtifact(artifact ArtifactInterface)
	IsEnableMavenDownloadOutput() bool
	SetEnableMavenDownloadOutput(enableMavenDownloadOutput bool)
}

// KogitoBuildStatusInterface ...
type KogitoBuildStatusInterface interface {
	GetLatestBuild() string
	SetLatestBuild(latestBuild string)
	GetConditions() []KogitoBuildConditionsInterface
	SetConditions(conditions []KogitoBuildConditionsInterface)
	AddCondition(condition KogitoBuildConditionsInterface)
	GetBuilds() BuildsInterface
	SetBuilds(builds BuildsInterface)
}

// BuildsInterface ...
type BuildsInterface interface {
	GetNew() []string
	SetNew(newBuilds []string)
	GetPending() []string
	SetPending(pendingBuilds []string)
	GetRunning() []string
	SetRunning(runningBuilds []string)
	GetComplete() []string
	SetComplete(completeBuilds []string)
	GetFailed() []string
	SetFailed(failedBuilds []string)
	GetError() []string
	SetError(errorBuilds []string)
	GetCancelled() []string
	SetCancelled(cancelled []string)
}

// KogitoBuildConditionsInterface ...
type KogitoBuildConditionsInterface interface {
	GetType() KogitoBuildConditionType
	SetType(conditionType KogitoBuildConditionType)
	GetStatus() corev1.ConditionStatus
	SetStatus(status corev1.ConditionStatus)
	GetLastTransitionTime() metav1.Time
	SetLastTransitionTime(lastTransitionTime metav1.Time)
	GetReason() KogitoBuildConditionReason
	SetReason(reason KogitoBuildConditionReason)
	GetMessage() string
	SetMessage(message string)
}

// KogitoBuildHandler ...
type KogitoBuildHandler interface {
	FetchKogitoBuildInstance(key types.NamespacedName) (KogitoBuildInterface, error)
}
