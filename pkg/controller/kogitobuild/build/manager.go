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

package build

import (
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	buildv1 "github.com/openshift/api/build/v1"
	imgv1 "github.com/openshift/api/image/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
)

const errorPrefix = "error while creating build resources: "

// Manager describes the interface to communicate with the build package.
// The controller typically manipulates the Kubernetes resources through this implementation.
type Manager interface {
	// GetRequestedResources gets the requested resources for the given KogitoBuild instance
	GetRequestedResources() (map[reflect.Type][]resource.KubernetesResource, error)
	// GetDeployedResources gets the deployed resources for the given KogitoBuild instance
	GetDeployedResources() (map[reflect.Type][]resource.KubernetesResource, error)
	// GetComparator gets the comparator to handle the equality logic between the requested and deployed resources
	GetComparator() compare.MapComparator
}

type manager struct {
	kogitoBuild *v1alpha1.KogitoBuild
	client      *client.Client
	scheme      *runtime.Scheme
}

func (m *manager) GetDeployedResources() (map[reflect.Type][]resource.KubernetesResource, error) {
	objectTypes := []runtime.Object{&buildv1.BuildConfigList{}}
	objectTypes = append(objectTypes, &imgv1.ImageStreamList{})
	resources, err := kubernetes.ResourceC(m.client).ListAll(objectTypes, m.kogitoBuild.Namespace, m.kogitoBuild)
	if err != nil {
		return nil, err
	}
	if err := services.AddSharedImageStreamToResources(resources, GetApplicationName(m.kogitoBuild), m.kogitoBuild.Namespace, m.client); err != nil {
		return nil, err
	}
	return resources, nil
}

func (m *manager) GetComparator() compare.MapComparator {
	resourceComparator := compare.DefaultComparator()
	resourceComparator.SetComparator(
		framework.NewComparatorBuilder().
			WithType(reflect.TypeOf(buildv1.BuildConfig{})).
			UseDefaultComparator().
			WithCustomComparator(framework.CreateBuildConfigComparator()).
			Build())

	resourceComparator.SetComparator(
		framework.NewComparatorBuilder().
			WithType(reflect.TypeOf(imgv1.ImageStream{})).
			UseDefaultComparator().
			WithCustomComparator(framework.CreateSharedImageStreamComparator()).
			Build())
	return compare.MapComparator{Comparator: resourceComparator}
}

// New creates a new Manager instance for the given KogitoBuild
func New(build *v1alpha1.KogitoBuild, client *client.Client, scheme *runtime.Scheme) (Manager, error) {
	setDefaults(build)
	if err := sanityCheck(build); err != nil {
		return nil, err
	}
	manager := manager{kogitoBuild: build, client: client, scheme: scheme}
	if v1alpha1.LocalSourceBuildType == build.Spec.Type ||
		v1alpha1.RemoteSourceBuildType == build.Spec.Type {
		return &sourceManager{manager}, nil
	}
	return &binaryManager{manager}, nil
}

// setDefaults sets the default values for the given KogitoBuild
func setDefaults(build *v1alpha1.KogitoBuild) {
	if len(build.Spec.Runtime) == 0 {
		build.Spec.Runtime = v1alpha1.QuarkusRuntimeType
	}
}

// sanityCheck verifies the spec attributes for the given KogitoBuild instance
func sanityCheck(build *v1alpha1.KogitoBuild) error {
	if len(build.Spec.Type) == 0 {
		return fmt.Errorf("%s: %s", errorPrefix, "build Type is required")
	}
	if build.Spec.Type == v1alpha1.RemoteSourceBuildType &&
		len(build.Spec.GitSource.URI) == 0 {
		return fmt.Errorf("%s: %s %s", errorPrefix, "Git URL is required when build type is", v1alpha1.RemoteSourceBuildType)
	}
	return nil
}
