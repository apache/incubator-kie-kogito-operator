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

package kogitoservice

import (
	api "github.com/kiegroup/kogito-operator/apis"
	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/manager"
	"github.com/kiegroup/kogito-operator/core/operator"
)

type runtimeDeployer struct {
	operator.Context
	definition   ServiceDefinition
	instance     api.KogitoService
	infraHandler manager.KogitoInfraHandler
}

// NewRuntimeDeployer creates a new ServiceDeployer to handle a custom Kogito Service instance to be handled by Operator SDK controller.
func NewRuntimeDeployer(context operator.Context, definition ServiceDefinition, instance api.KogitoService, infraHandler manager.KogitoInfraHandler) ServiceDeployer {
	if len(definition.Request.NamespacedName.Namespace) == 0 && len(definition.Request.NamespacedName.Name) == 0 {
		panic("No Request provided for the Service Deployer")
	}
	return &runtimeDeployer{
		Context:      context,
		definition:   definition,
		instance:     instance,
		infraHandler: infraHandler,
	}
}

func (s *runtimeDeployer) Deploy() error {
	if s.instance.GetSpec().GetReplicas() == nil {
		s.instance.GetSpec().SetReplicas(defaultReplicas)
	}
	if len(s.definition.DefaultImageName) == 0 {
		s.definition.DefaultImageName = s.definition.Request.Name
	}

	var err error

	// always updateStatus its status
	statusHandler := NewStatusHandler(s.Context)
	defer statusHandler.HandleStatusUpdate(s.instance, &err)

	s.definition.Envs = s.instance.GetSpec().GetEnvs()

	infraPropertiesReconciler := newConfigReconciler(s.Context, s.instance, &s.definition)
	if err = infraPropertiesReconciler.Reconcile(); err != nil {
		return err
	}

	configMapReferenceReconciler := newPropertiesConfigMapReconciler(s.Context, s.instance, &s.definition)
	if err = configMapReferenceReconciler.Reconcile(); err != nil {
		return err
	}

	trustStoreReconciler := newTrustStoreReconciler(s.Context, s.instance, &s.definition)
	if err = trustStoreReconciler.Reconcile(); err != nil {
		return err
	}

	kogitoInfraReconciler := newKogitoInfraReconciler(s.Context, s.instance, &s.definition, s.infraHandler)
	if err = kogitoInfraReconciler.Reconcile(); err != nil {
		return err
	}

	imageHandler := s.newImageHandler()
	if err = imageHandler.ReconcileImageStream(s.instance); err != nil {
		return err
	}

	deploymentReconciler := newDeploymentReconciler(s.Context, s.instance, s.definition, imageHandler)
	if err = deploymentReconciler.Reconcile(); err != nil {
		return err
	}

	return err
}

func (s *runtimeDeployer) newImageHandler() infrastructure.ImageHandler {
	addDockerImageReference := len(s.instance.GetSpec().GetImage()) != 0 || !s.definition.CustomService
	image := s.resolveImage()
	return infrastructure.NewImageHandler(s.Context, image, s.definition.DefaultImageName, image.Name, s.instance.GetNamespace(), addDockerImageReference, s.instance.GetSpec().IsInsecureImageRegistry())
}

func (s *runtimeDeployer) resolveImage() *api.Image {
	var image api.Image
	if len(s.instance.GetSpec().GetImage()) == 0 {
		image = api.Image{
			Name: s.definition.DefaultImageName,
			Tag:  s.definition.DefaultImageTag,
		}
	} else {
		image = framework.ConvertImageTagToImage(s.instance.GetSpec().GetImage())
	}
	return &image
}
