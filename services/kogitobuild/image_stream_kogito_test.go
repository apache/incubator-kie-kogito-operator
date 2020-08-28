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

package kogitobuild

import (
	"github.com/kiegroup/kogito-cloud-operator/api/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_resolveSourceStrategyImageNameForBuilds(t *testing.T) {
	buildQuarkusNonNative := &v1alpha1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{Name: "buildQuarkusNonNative", Namespace: t.Name()},
		Spec:       v1alpha1.KogitoBuildSpec{Runtime: v1alpha1.QuarkusRuntimeType, Native: false},
	}
	buildQuarkusNative := &v1alpha1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{Name: "buildQuarkusNative", Namespace: t.Name()},
		Spec:       v1alpha1.KogitoBuildSpec{Runtime: v1alpha1.QuarkusRuntimeType, Native: true},
	}
	buildSpringBoot := &v1alpha1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{Name: "buildSpringBoot", Namespace: t.Name()},
		Spec:       v1alpha1.KogitoBuildSpec{Runtime: v1alpha1.SpringBootRuntimeType},
	}
	buildQuarkusCustom := &v1alpha1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{Name: "buildQuarkusCustom", Namespace: t.Name()},
		Spec:       v1alpha1.KogitoBuildSpec{Runtime: v1alpha1.QuarkusRuntimeType, RuntimeImage: v1alpha1.Image{Name: "my-image"}},
	}
	buildSpringBootCustom := &v1alpha1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{Name: "buildSpringBootCustom", Namespace: t.Name()},
		Spec:       v1alpha1.KogitoBuildSpec{Runtime: v1alpha1.SpringBootRuntimeType, BuildImage: v1alpha1.Image{Name: "my-image"}},
	}
	tag := ":" + infrastructure.GetKogitoImageVersion()
	type args struct {
		build        *v1alpha1.KogitoBuild
		builderImage bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Quarkus Non Native Builder", args{buildQuarkusNonNative, true}, KogitoQuarkusUbi8s2iImage + tag},
		{"Quarkus Non Native Base", args{buildQuarkusNonNative, false}, KogitoQuarkusJVMUbi8Image + tag},
		{"Quarkus Native Builder", args{buildQuarkusNative, true}, KogitoQuarkusUbi8s2iImage + tag},
		{"Quarkus Native Base", args{buildQuarkusNative, false}, KogitoQuarkusUbi8Image + tag},
		{"SpringBoot Builder", args{buildSpringBoot, true}, KogitoSpringBootUbi8s2iImage + tag},
		{"SpringBoot Base", args{buildSpringBoot, false}, KogitoSpringBootUbi8Image + tag},
		{"SpringBoot Custom Builder", args{buildSpringBootCustom, true}, "my-image" + tag},
		{"Quarkus Custom Base", args{buildQuarkusCustom, false}, "my-image" + tag},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveKogitoImageNameTag(tt.args.build, tt.args.builderImage); got != tt.want {
				t.Errorf("resolveKogitoImageNameTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolveKogitoImageStreamName(t *testing.T) {
	buildQuarkusNonNative := &v1alpha1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{Name: "buildQuarkusNonNative", Namespace: t.Name()},
		Spec:       v1alpha1.KogitoBuildSpec{Runtime: v1alpha1.QuarkusRuntimeType, Native: false},
	}
	buildQuarkusNative := &v1alpha1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{Name: "buildQuarkusNative", Namespace: t.Name()},
		Spec:       v1alpha1.KogitoBuildSpec{Runtime: v1alpha1.QuarkusRuntimeType, Native: true},
	}
	buildSpringBoot := &v1alpha1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{Name: "buildSpringBoot", Namespace: t.Name()},
		Spec:       v1alpha1.KogitoBuildSpec{Runtime: v1alpha1.SpringBootRuntimeType},
	}
	buildQuarkusCustom := &v1alpha1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{Name: "buildQuarkusCustom", Namespace: t.Name()},
		Spec:       v1alpha1.KogitoBuildSpec{Runtime: v1alpha1.QuarkusRuntimeType, RuntimeImage: v1alpha1.Image{Name: "my-image"}},
	}
	buildSpringBootCustom := &v1alpha1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{Name: "buildSpringBootCustom", Namespace: t.Name()},
		Spec:       v1alpha1.KogitoBuildSpec{Runtime: v1alpha1.SpringBootRuntimeType, BuildImage: v1alpha1.Image{Name: "my-image"}},
	}
	type args struct {
		build     *v1alpha1.KogitoBuild
		isBuilder bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Quarkus Non Native Builder", args{buildQuarkusNonNative, true}, KogitoQuarkusUbi8s2iImage},
		{"Quarkus Non Native Base", args{buildQuarkusNonNative, false}, KogitoQuarkusJVMUbi8Image},
		{"Quarkus Native Builder", args{buildQuarkusNative, true}, KogitoQuarkusUbi8s2iImage},
		{"Quarkus Native Base", args{buildQuarkusNative, false}, KogitoQuarkusUbi8Image},
		{"SpringBoot Builder", args{buildSpringBoot, true}, KogitoSpringBootUbi8s2iImage},
		{"SpringBoot Base", args{buildSpringBoot, false}, KogitoSpringBootUbi8Image},
		{"SpringBoot Custom Builder", args{buildSpringBootCustom, true}, "custom-my-image"},
		{"Quarkus Custom Base", args{buildQuarkusCustom, false}, "custom-my-image"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveKogitoImageStreamName(tt.args.build, tt.args.isBuilder); got != tt.want {
				t.Errorf("resolveKogitoImageStreamName() = %v, want %v", got, tt.want)
			}
		})
	}
}
