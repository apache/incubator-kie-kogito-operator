// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Artifact) DeepCopyInto(out *Artifact) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Artifact.
func (in *Artifact) DeepCopy() *Artifact {
	if in == nil {
		return nil
	}
	out := new(Artifact)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Builds) DeepCopyInto(out *Builds) {
	*out = *in
	if in.New != nil {
		in, out := &in.New, &out.New
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Pending != nil {
		in, out := &in.Pending, &out.Pending
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Running != nil {
		in, out := &in.Running, &out.Running
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Complete != nil {
		in, out := &in.Complete, &out.Complete
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Failed != nil {
		in, out := &in.Failed, &out.Failed
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Error != nil {
		in, out := &in.Error, &out.Error
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Cancelled != nil {
		in, out := &in.Cancelled, &out.Cancelled
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Builds.
func (in *Builds) DeepCopy() *Builds {
	if in == nil {
		return nil
	}
	out := new(Builds)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitSource) DeepCopyInto(out *GitSource) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitSource.
func (in *GitSource) DeepCopy() *GitSource {
	if in == nil {
		return nil
	}
	out := new(GitSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InfraResource) DeepCopyInto(out *InfraResource) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InfraResource.
func (in *InfraResource) DeepCopy() *InfraResource {
	if in == nil {
		return nil
	}
	out := new(InfraResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoBuild) DeepCopyInto(out *KogitoBuild) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoBuild.
func (in *KogitoBuild) DeepCopy() *KogitoBuild {
	if in == nil {
		return nil
	}
	out := new(KogitoBuild)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KogitoBuild) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoBuildList) DeepCopyInto(out *KogitoBuildList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KogitoBuild, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoBuildList.
func (in *KogitoBuildList) DeepCopy() *KogitoBuildList {
	if in == nil {
		return nil
	}
	out := new(KogitoBuildList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KogitoBuildList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoBuildSpec) DeepCopyInto(out *KogitoBuildSpec) {
	*out = *in
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]corev1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.GitSource = in.GitSource
	if in.WebHooks != nil {
		in, out := &in.WebHooks, &out.WebHooks
		*out = make([]WebHookSecret, len(*in))
		copy(*out, *in)
	}
	in.Resources.DeepCopyInto(&out.Resources)
	out.Artifact = in.Artifact
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoBuildSpec.
func (in *KogitoBuildSpec) DeepCopy() *KogitoBuildSpec {
	if in == nil {
		return nil
	}
	out := new(KogitoBuildSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoBuildStatus) DeepCopyInto(out *KogitoBuildStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = new([]metav1.Condition)
		if **in != nil {
			in, out := *in, *out
			*out = make([]metav1.Condition, len(*in))
			for i := range *in {
				(*in)[i].DeepCopyInto(&(*out)[i])
			}
		}
	}
	in.Builds.DeepCopyInto(&out.Builds)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoBuildStatus.
func (in *KogitoBuildStatus) DeepCopy() *KogitoBuildStatus {
	if in == nil {
		return nil
	}
	out := new(KogitoBuildStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoCloudEventInfo) DeepCopyInto(out *KogitoCloudEventInfo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoCloudEventInfo.
func (in *KogitoCloudEventInfo) DeepCopy() *KogitoCloudEventInfo {
	if in == nil {
		return nil
	}
	out := new(KogitoCloudEventInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoCloudEventsStatus) DeepCopyInto(out *KogitoCloudEventsStatus) {
	*out = *in
	if in.Consumes != nil {
		in, out := &in.Consumes, &out.Consumes
		*out = make([]KogitoCloudEventInfo, len(*in))
		copy(*out, *in)
	}
	if in.Produces != nil {
		in, out := &in.Produces, &out.Produces
		*out = make([]KogitoCloudEventInfo, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoCloudEventsStatus.
func (in *KogitoCloudEventsStatus) DeepCopy() *KogitoCloudEventsStatus {
	if in == nil {
		return nil
	}
	out := new(KogitoCloudEventsStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoInfra) DeepCopyInto(out *KogitoInfra) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoInfra.
func (in *KogitoInfra) DeepCopy() *KogitoInfra {
	if in == nil {
		return nil
	}
	out := new(KogitoInfra)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KogitoInfra) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoInfraList) DeepCopyInto(out *KogitoInfraList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KogitoInfra, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoInfraList.
func (in *KogitoInfraList) DeepCopy() *KogitoInfraList {
	if in == nil {
		return nil
	}
	out := new(KogitoInfraList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KogitoInfraList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoInfraSpec) DeepCopyInto(out *KogitoInfraSpec) {
	*out = *in
	if in.Resource != nil {
		in, out := &in.Resource, &out.Resource
		*out = new(InfraResource)
		**out = **in
	}
	if in.InfraProperties != nil {
		in, out := &in.InfraProperties, &out.InfraProperties
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Envs != nil {
		in, out := &in.Envs, &out.Envs
		*out = make([]corev1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ConfigMapEnvFromReferences != nil {
		in, out := &in.ConfigMapEnvFromReferences, &out.ConfigMapEnvFromReferences
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ConfigMapVolumeReferences != nil {
		in, out := &in.ConfigMapVolumeReferences, &out.ConfigMapVolumeReferences
		*out = make([]VolumeReference, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SecretEnvFromReferences != nil {
		in, out := &in.SecretEnvFromReferences, &out.SecretEnvFromReferences
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SecretVolumeReferences != nil {
		in, out := &in.SecretVolumeReferences, &out.SecretVolumeReferences
		*out = make([]VolumeReference, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoInfraSpec.
func (in *KogitoInfraSpec) DeepCopy() *KogitoInfraSpec {
	if in == nil {
		return nil
	}
	out := new(KogitoInfraSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoInfraStatus) DeepCopyInto(out *KogitoInfraStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = new([]metav1.Condition)
		if **in != nil {
			in, out := *in, *out
			*out = make([]metav1.Condition, len(*in))
			for i := range *in {
				(*in)[i].DeepCopyInto(&(*out)[i])
			}
		}
	}
	if in.Envs != nil {
		in, out := &in.Envs, &out.Envs
		*out = make([]corev1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ConfigMapEnvFromReferences != nil {
		in, out := &in.ConfigMapEnvFromReferences, &out.ConfigMapEnvFromReferences
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ConfigMapVolumeReferences != nil {
		in, out := &in.ConfigMapVolumeReferences, &out.ConfigMapVolumeReferences
		*out = make([]VolumeReference, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SecretEnvFromReferences != nil {
		in, out := &in.SecretEnvFromReferences, &out.SecretEnvFromReferences
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SecretVolumeReferences != nil {
		in, out := &in.SecretVolumeReferences, &out.SecretVolumeReferences
		*out = make([]VolumeReference, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoInfraStatus.
func (in *KogitoInfraStatus) DeepCopy() *KogitoInfraStatus {
	if in == nil {
		return nil
	}
	out := new(KogitoInfraStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoProbe) DeepCopyInto(out *KogitoProbe) {
	*out = *in
	in.LivenessProbe.DeepCopyInto(&out.LivenessProbe)
	in.ReadinessProbe.DeepCopyInto(&out.ReadinessProbe)
	in.StartupProbe.DeepCopyInto(&out.StartupProbe)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoProbe.
func (in *KogitoProbe) DeepCopy() *KogitoProbe {
	if in == nil {
		return nil
	}
	out := new(KogitoProbe)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoRuntime) DeepCopyInto(out *KogitoRuntime) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoRuntime.
func (in *KogitoRuntime) DeepCopy() *KogitoRuntime {
	if in == nil {
		return nil
	}
	out := new(KogitoRuntime)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KogitoRuntime) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoRuntimeList) DeepCopyInto(out *KogitoRuntimeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KogitoRuntime, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoRuntimeList.
func (in *KogitoRuntimeList) DeepCopy() *KogitoRuntimeList {
	if in == nil {
		return nil
	}
	out := new(KogitoRuntimeList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KogitoRuntimeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoRuntimeSpec) DeepCopyInto(out *KogitoRuntimeSpec) {
	*out = *in
	in.KogitoServiceSpec.DeepCopyInto(&out.KogitoServiceSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoRuntimeSpec.
func (in *KogitoRuntimeSpec) DeepCopy() *KogitoRuntimeSpec {
	if in == nil {
		return nil
	}
	out := new(KogitoRuntimeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoRuntimeStatus) DeepCopyInto(out *KogitoRuntimeStatus) {
	*out = *in
	in.KogitoServiceStatus.DeepCopyInto(&out.KogitoServiceStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoRuntimeStatus.
func (in *KogitoRuntimeStatus) DeepCopy() *KogitoRuntimeStatus {
	if in == nil {
		return nil
	}
	out := new(KogitoRuntimeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoServiceSpec) DeepCopyInto(out *KogitoServiceSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]corev1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Resources.DeepCopyInto(&out.Resources)
	if in.DeploymentLabels != nil {
		in, out := &in.DeploymentLabels, &out.DeploymentLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.ServiceLabels != nil {
		in, out := &in.ServiceLabels, &out.ServiceLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Infra != nil {
		in, out := &in.Infra, &out.Infra
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	out.Monitoring = in.Monitoring
	if in.Config != nil {
		in, out := &in.Config, &out.Config
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.Probes.DeepCopyInto(&out.Probes)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoServiceSpec.
func (in *KogitoServiceSpec) DeepCopy() *KogitoServiceSpec {
	if in == nil {
		return nil
	}
	out := new(KogitoServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoServiceStatus) DeepCopyInto(out *KogitoServiceStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = new([]metav1.Condition)
		if **in != nil {
			in, out := *in, *out
			*out = make([]metav1.Condition, len(*in))
			for i := range *in {
				(*in)[i].DeepCopyInto(&(*out)[i])
			}
		}
	}
	if in.DeploymentConditions != nil {
		in, out := &in.DeploymentConditions, &out.DeploymentConditions
		*out = make([]appsv1.DeploymentCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.RouteConditions != nil {
		in, out := &in.RouteConditions, &out.RouteConditions
		*out = new([]metav1.Condition)
		if **in != nil {
			in, out := *in, *out
			*out = make([]metav1.Condition, len(*in))
			for i := range *in {
				(*in)[i].DeepCopyInto(&(*out)[i])
			}
		}
	}
	in.CloudEvents.DeepCopyInto(&out.CloudEvents)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoServiceStatus.
func (in *KogitoServiceStatus) DeepCopy() *KogitoServiceStatus {
	if in == nil {
		return nil
	}
	out := new(KogitoServiceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoSupportingService) DeepCopyInto(out *KogitoSupportingService) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoSupportingService.
func (in *KogitoSupportingService) DeepCopy() *KogitoSupportingService {
	if in == nil {
		return nil
	}
	out := new(KogitoSupportingService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KogitoSupportingService) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoSupportingServiceList) DeepCopyInto(out *KogitoSupportingServiceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KogitoSupportingService, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoSupportingServiceList.
func (in *KogitoSupportingServiceList) DeepCopy() *KogitoSupportingServiceList {
	if in == nil {
		return nil
	}
	out := new(KogitoSupportingServiceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KogitoSupportingServiceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoSupportingServiceSpec) DeepCopyInto(out *KogitoSupportingServiceSpec) {
	*out = *in
	in.KogitoServiceSpec.DeepCopyInto(&out.KogitoServiceSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoSupportingServiceSpec.
func (in *KogitoSupportingServiceSpec) DeepCopy() *KogitoSupportingServiceSpec {
	if in == nil {
		return nil
	}
	out := new(KogitoSupportingServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KogitoSupportingServiceStatus) DeepCopyInto(out *KogitoSupportingServiceStatus) {
	*out = *in
	in.KogitoServiceStatus.DeepCopyInto(&out.KogitoServiceStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KogitoSupportingServiceStatus.
func (in *KogitoSupportingServiceStatus) DeepCopy() *KogitoSupportingServiceStatus {
	if in == nil {
		return nil
	}
	out := new(KogitoSupportingServiceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Monitoring) DeepCopyInto(out *Monitoring) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Monitoring.
func (in *Monitoring) DeepCopy() *Monitoring {
	if in == nil {
		return nil
	}
	out := new(Monitoring)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeReference) DeepCopyInto(out *VolumeReference) {
	*out = *in
	if in.FileMode != nil {
		in, out := &in.FileMode, &out.FileMode
		*out = new(int32)
		**out = **in
	}
	if in.Optional != nil {
		in, out := &in.Optional, &out.Optional
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeReference.
func (in *VolumeReference) DeepCopy() *VolumeReference {
	if in == nil {
		return nil
	}
	out := new(VolumeReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebHookSecret) DeepCopyInto(out *WebHookSecret) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebHookSecret.
func (in *WebHookSecret) DeepCopy() *WebHookSecret {
	if in == nil {
		return nil
	}
	out := new(WebHookSecret)
	in.DeepCopyInto(out)
	return out
}
