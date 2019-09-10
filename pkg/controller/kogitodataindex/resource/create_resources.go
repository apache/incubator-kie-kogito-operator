package resource

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/resource"
	routev1 "github.com/openshift/api/route/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var log = logger.GetLogger("resources_kogitodataindex")

// KogitoDataIndexResources is the data structure for all resources managed by KogitoDataIndex CR
type KogitoDataIndexResources struct {
	// StatefulSet is the resource responsible for the Data Index Service image deployment in the cluster
	StatefulSet       *appsv1.StatefulSet
	StatefulSetStatus KogitoDataIndexResourcesStatus
	// ProtoBufConfigMap will mount a config map for protobuf files that the Data Index Service will read
	ProtoBufConfigMap       *corev1.ConfigMap
	ProtoBufConfigMapStatus KogitoDataIndexResourcesStatus
	// Service to expose the data index service internally
	Service       *corev1.Service
	ServiceStatus KogitoDataIndexResourcesStatus
	// Route to expose the service in the Ingress. Supported only on OpenShift for now
	Route       *routev1.Route
	RouteStatus KogitoDataIndexResourcesStatus
}

// KogitoDataIndexResourcesStatus identifies the status of the resource
type KogitoDataIndexResourcesStatus struct {
	New bool
}

type kogitoDataIndexResourcesFactory struct {
	resource.Factory
	Resources       *KogitoDataIndexResources
	KogitoDataIndex *v1alpha1.KogitoDataIndex
}

// Build will call a builder function if no errors were found
func (f *kogitoDataIndexResourcesFactory) build(fn func(*kogitoDataIndexResourcesFactory) *kogitoDataIndexResourcesFactory) *kogitoDataIndexResourcesFactory {
	if f.Error == nil {
		return fn(f)
	}
	// break the chain
	return f
}

func (f *kogitoDataIndexResourcesFactory) buildOnOpenshift(fn func(*kogitoDataIndexResourcesFactory) *kogitoDataIndexResourcesFactory) *kogitoDataIndexResourcesFactory {
	if f.Error == nil && f.Context.Client.IsOpenshift() {
		return fn(f)
	}
	// break the chain
	return f
}

// CreateOrFetchResources will create the needed resources in the cluster if they not exists, fetch otherwise
func CreateOrFetchResources(instance *v1alpha1.KogitoDataIndex, context resource.FactoryContext) (KogitoDataIndexResources, error) {
	factory := kogitoDataIndexResourcesFactory{
		Factory:         resource.Factory{Context: &context},
		Resources:       &KogitoDataIndexResources{},
		KogitoDataIndex: instance,
	}

	// TODO: add Kafka, KafkaTopic and Infinispan
	factory.
		build(createProtoBufConfigMap).
		build(createStatefulSet).
		build(createService).
		buildOnOpenshift(createRoute)

	return *factory.Resources, factory.Error
}

func createProtoBufConfigMap(f *kogitoDataIndexResourcesFactory) *kogitoDataIndexResourcesFactory {
	cm := newProtobufConfigMap(f.KogitoDataIndex)
	if err := f.CallPreCreate(cm); err != nil {
		f.Error = err
		return f
	}

	if f.Resources.ProtoBufConfigMapStatus.New, f.Error =
		kubernetes.ResourceC(f.Context.Client).CreateIfNotExists(cm); f.Error != nil {
		return f
	}

	if f.CallPostCreate(f.Resources.ProtoBufConfigMapStatus.New, cm); f.Error != nil {
		return f
	}

	f.Resources.ProtoBufConfigMap = cm

	return f
}

func createStatefulSet(f *kogitoDataIndexResourcesFactory) *kogitoDataIndexResourcesFactory {
	secret, err := fetchInfinispanCredentials(f.KogitoDataIndex, f.Context.Client)
	if err != nil {
		f.Error = err
		return f
	}
	statefulset := newStatefulset(f.KogitoDataIndex, f.Resources.ProtoBufConfigMap, *secret)
	if err := f.CallPreCreate(statefulset); err != nil {
		f.Error = err
		return f
	}

	if f.Resources.StatefulSetStatus.New, f.Error =
		kubernetes.ResourceC(f.Context.Client).CreateIfNotExists(statefulset); f.Error != nil {
		return f
	}

	if f.CallPostCreate(f.Resources.StatefulSetStatus.New, statefulset); f.Error != nil {
		return f
	}

	f.Resources.StatefulSet = statefulset

	return f
}

func createService(f *kogitoDataIndexResourcesFactory) *kogitoDataIndexResourcesFactory {
	svc := newService(f.KogitoDataIndex, f.Resources.StatefulSet)
	if f.Error = f.CallPreCreate(svc); f.Error != nil {
		return f
	}

	if f.Resources.ServiceStatus.New, f.Error =
		kubernetes.ResourceC(f.Context.Client).CreateIfNotExists(svc); f.Error != nil {
		return f
	}

	if f.CallPostCreate(f.Resources.ServiceStatus.New, svc); f.Error != nil {
		return f
	}

	f.Resources.Service = svc

	return f
}

func createRoute(f *kogitoDataIndexResourcesFactory) *kogitoDataIndexResourcesFactory {
	route, err := newRoute(f.KogitoDataIndex, f.Resources.Service)
	if err != nil {
		f.Error = err
		return f
	}

	if f.Error = f.CallPreCreate(route); f.Error != nil {
		return f
	}

	if f.Resources.RouteStatus.New, f.Error = kubernetes.ResourceC(f.Context.Client).CreateIfNotExists(route); f.Error != nil {
		return f
	}

	if f.CallPostCreate(f.Resources.RouteStatus.New, route); f.Error != nil {
		return f
	}

	f.Resources.Route = route

	return f
}
