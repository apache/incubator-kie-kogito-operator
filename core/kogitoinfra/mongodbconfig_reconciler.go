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

package kogitoinfra

import (
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/api/v1beta1"
	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	mongodb "github.com/mongodb/mongodb-kubernetes-operator/pkg/apis/mongodb/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/url"
	"reflect"
)

const (
	mongoDBConfigMapName = "kogito-mongodb-%s-config"
)

type mongoDBConfigReconciler struct {
	infraContext
	mongoDBInstance  *mongodb.MongoDB
	runtime          api.RuntimeType
	configMapHandler infrastructure.ConfigMapHandler
}

func newMongoDBConfigReconciler(ctx infraContext, mongoDBInstance *mongodb.MongoDB, runtime api.RuntimeType) Reconciler {
	return &mongoDBConfigReconciler{
		infraContext:     ctx,
		mongoDBInstance:  mongoDBInstance,
		runtime:          runtime,
		configMapHandler: infrastructure.NewConfigMapHandler(ctx.Context),
	}
}

func (i *mongoDBConfigReconciler) Reconcile() (err error) {

	// Create Required resource
	requestedResources, err := i.createRequiredResources()
	if err != nil {
		return
	}

	// Get Deployed resource
	deployedResources, err := i.getDeployedResources()
	if err != nil {
		return
	}

	// Process Delta
	if err = i.processDelta(requestedResources, deployedResources); err != nil {
		return err
	}

	configMapReference := &v1beta1.ConfigMapReference{
		Name:      i.getMongoDBConfigMapName(),
		MountType: api.EnvVar,
	}
	i.updateConfigMapReferenceInStatus(configMapReference)
	return nil
}

func (i *mongoDBConfigReconciler) createRequiredResources() (map[reflect.Type][]resource.KubernetesResource, error) {
	resources := make(map[reflect.Type][]resource.KubernetesResource)
	appProps, err := i.getMongoDBAppProps()
	if err != nil {
		return nil, err
	}
	configMap := i.createMongoDBConfigMap(appProps)
	if err := framework.SetOwner(i.infraContext.instance, i.infraContext.Scheme, configMap); err != nil {
		return resources, err
	}
	resources[reflect.TypeOf(v12.ConfigMap{})] = []resource.KubernetesResource{configMap}
	return resources, nil
}

func (i *mongoDBConfigReconciler) getDeployedResources() (map[reflect.Type][]resource.KubernetesResource, error) {
	resources := make(map[reflect.Type][]resource.KubernetesResource)
	// fetch owned image stream
	deployedConfigMap, err := i.configMapHandler.FetchConfigMap(types.NamespacedName{Name: i.getMongoDBConfigMapName(), Namespace: i.infraContext.instance.GetNamespace()})
	if err != nil {
		return nil, err
	}
	if deployedConfigMap != nil {
		resources[reflect.TypeOf(v12.ConfigMap{})] = []resource.KubernetesResource{deployedConfigMap}
	}
	return resources, nil
}

func (i *mongoDBConfigReconciler) processDelta(requestedResources map[reflect.Type][]resource.KubernetesResource, deployedResources map[reflect.Type][]resource.KubernetesResource) (err error) {
	comparator := i.configMapHandler.GetComparator()
	deltaProcessor := infrastructure.NewDeltaProcessor(i.infraContext.Context)
	_, err = deltaProcessor.ProcessDelta(comparator, requestedResources, deployedResources)
	return err
}

func (i *mongoDBConfigReconciler) getMongoDBAppProps() (map[string]string, error) {
	appProps := map[string]string{}

	mongoDBURI := i.mongoDBInstance.Status.MongoURI
	if len(mongoDBURI) > 0 {
		appProps[mongoDBEnablePersistenceEnvKey] = "true"
		mongoDBParsedURL, err := url.ParseRequestURI(mongoDBURI)
		if err != nil {
			return nil, err
		}
		if i.runtime == api.QuarkusRuntimeType {
			appProps[propertiesMongoDB[i.runtime][appPropMongoDBURI]] = mongoDBURI
		} else if i.runtime == api.SpringBootRuntimeType {
			appProps[propertiesMongoDB[i.runtime][appPropMongoDBHost]] = mongoDBParsedURL.Hostname()
			appProps[propertiesMongoDB[i.runtime][appPropMongoDBPort]] = mongoDBParsedURL.Port()
		}
	}
	return appProps, nil
}

func (i *mongoDBConfigReconciler) createMongoDBConfigMap(appProps map[string]string) *v12.ConfigMap {
	var data map[string]string = nil
	if len(appProps) > 0 {
		data = appProps
	}
	configMap := &v12.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      i.getMongoDBConfigMapName(),
			Namespace: i.infraContext.instance.GetNamespace(),
			Labels: map[string]string{
				framework.LabelAppKey: i.infraContext.instance.GetName(),
			},
		},
		Data: data,
	}
	return configMap
}

func (i *mongoDBConfigReconciler) getMongoDBConfigMapName() string {
	return fmt.Sprintf(mongoDBConfigMapName, i.runtime)
}

func (i *mongoDBConfigReconciler) updateConfigMapReferenceInStatus(configMapReference *v1beta1.ConfigMapReference) {
	instance := i.infraContext.instance
	configMapReferences := append(instance.GetStatus().GetConfigMapReferences(), configMapReference)
	instance.GetStatus().SetConfigMapReferences(configMapReferences)
}
