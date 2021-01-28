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
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/infrastructure"
	api2 "github.com/kiegroup/kogito-cloud-operator/core/test/api"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_resolveSourceStrategyImageNameForBuilds(t *testing.T) {
	buildQuarkusNonNative := &api2.KogitoBuildTest{
		ObjectMeta: metav1.ObjectMeta{Name: "buildQuarkusNonNative", Namespace: t.Name()},
		Spec:       api.KogitoBuildSpec{Runtime: api.QuarkusRuntimeType, Native: false},
	}
	buildQuarkusNative := &api2.KogitoBuildTest{
		ObjectMeta: metav1.ObjectMeta{Name: "buildQuarkusNative", Namespace: t.Name()},
		Spec:       api.KogitoBuildSpec{Runtime: api.QuarkusRuntimeType, Native: true},
	}
	buildSpringBoot := &api2.KogitoBuildTest{
		ObjectMeta: metav1.ObjectMeta{Name: "buildSpringBoot", Namespace: t.Name()},
		Spec:       api.KogitoBuildSpec{Runtime: api.SpringBootRuntimeType},
	}
	buildQuarkusCustom := &api2.KogitoBuildTest{
		ObjectMeta: metav1.ObjectMeta{Name: "buildQuarkusCustom", Namespace: t.Name()},
		Spec:       api.KogitoBuildSpec{Runtime: api.QuarkusRuntimeType, RuntimeImage: "my-image:1.0"},
	}
	buildSpringBootCustom := &api2.KogitoBuildTest{
		ObjectMeta: metav1.ObjectMeta{Name: "buildSpringBootCustom", Namespace: t.Name()},
		Spec:       api.KogitoBuildSpec{Runtime: api.SpringBootRuntimeType, BuildImage: "my-image:1.0"},
	}
	tag := ":" + infrastructure.GetKogitoImageVersion()
	type args struct {
		build        *api2.KogitoBuildTest
		builderImage bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Quarkus Non Native Builder", args{buildQuarkusNonNative, true}, infrastructure.KogitoQuarkusUbi8s2iImage + tag},
		{"Quarkus Non Native Base", args{buildQuarkusNonNative, false}, infrastructure.KogitoQuarkusJVMUbi8Image + tag},
		{"Quarkus Native Builder", args{buildQuarkusNative, true}, infrastructure.KogitoQuarkusUbi8s2iImage + tag},
		{"Quarkus Native Base", args{buildQuarkusNative, false}, infrastructure.KogitoQuarkusUbi8Image + tag},
		{"SpringBoot Builder", args{buildSpringBoot, true}, infrastructure.KogitoSpringBootUbi8s2iImage + tag},
		{"SpringBoot Base", args{buildSpringBoot, false}, infrastructure.KogitoSpringBootUbi8Image + tag},
		{"SpringBoot Custom Builder", args{buildSpringBootCustom, true}, "my-image:1.0"},
		{"Quarkus Custom Base", args{buildQuarkusCustom, false}, "my-image:1.0"},
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
	buildQuarkusNonNative := &api2.KogitoBuildTest{
		ObjectMeta: metav1.ObjectMeta{Name: "buildQuarkusNonNative", Namespace: t.Name()},
		Spec:       api.KogitoBuildSpec{Runtime: api.QuarkusRuntimeType, Native: false},
	}
	buildQuarkusNative := &api2.KogitoBuildTest{
		ObjectMeta: metav1.ObjectMeta{Name: "buildQuarkusNative", Namespace: t.Name()},
		Spec:       api.KogitoBuildSpec{Runtime: api.QuarkusRuntimeType, Native: true},
	}
	buildSpringBoot := &api2.KogitoBuildTest{
		ObjectMeta: metav1.ObjectMeta{Name: "buildSpringBoot", Namespace: t.Name()},
		Spec:       api.KogitoBuildSpec{Runtime: api.SpringBootRuntimeType},
	}
	buildQuarkusCustom := &api2.KogitoBuildTest{
		ObjectMeta: metav1.ObjectMeta{Name: "buildQuarkusCustom", Namespace: t.Name()},
		Spec:       api.KogitoBuildSpec{Runtime: api.QuarkusRuntimeType, RuntimeImage: "my-image"},
	}
	buildSpringBootCustom := &api2.KogitoBuildTest{
		ObjectMeta: metav1.ObjectMeta{Name: "buildSpringBootCustom", Namespace: t.Name()},
		Spec:       api.KogitoBuildSpec{Runtime: api.SpringBootRuntimeType, BuildImage: "my-image"},
	}
	type args struct {
		build     *api2.KogitoBuildTest
		isBuilder bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Quarkus Non Native Builder", args{buildQuarkusNonNative, true}, infrastructure.KogitoQuarkusUbi8s2iImage},
		{"Quarkus Non Native Base", args{buildQuarkusNonNative, false}, infrastructure.KogitoQuarkusJVMUbi8Image},
		{"Quarkus Native Builder", args{buildQuarkusNative, true}, infrastructure.KogitoQuarkusUbi8s2iImage},
		{"Quarkus Native Base", args{buildQuarkusNative, false}, infrastructure.KogitoQuarkusUbi8Image},
		{"SpringBoot Builder", args{buildSpringBoot, true}, infrastructure.KogitoSpringBootUbi8s2iImage},
		{"SpringBoot Base", args{buildSpringBoot, false}, infrastructure.KogitoSpringBootUbi8Image},
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
