package manager

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"k8s.io/api/apps/v1"
)

// KogitoRuntimeManager ...
type KogitoRuntimeManager interface {
	FetchKogitoRuntimeDeployments(namespace string) ([]v1.Deployment, error)
}

type kogitoRuntimeManager struct {
	client         *client.Client
	log            logger.Logger
	runtimeHandler api.KogitoRuntimeHandler
}

// NewKogitoRuntimeManager ...
func NewKogitoRuntimeManager(client *client.Client, log logger.Logger, runtimeHandler api.KogitoRuntimeHandler) KogitoRuntimeManager {
	return &kogitoRuntimeManager{
		client:         client,
		log:            log,
		runtimeHandler: runtimeHandler,
	}
}

// FetchKogitoRuntimeDeployments gets all dcs owned by KogitoRuntime services within the given namespace
func (k *kogitoRuntimeManager) FetchKogitoRuntimeDeployments(namespace string) ([]v1.Deployment, error) {
	var kdcs []v1.Deployment
	kogitoRuntimeServices, err := k.runtimeHandler.FetchAllKogitoRuntimeInstances(namespace)
	if err != nil {
		return nil, err
	} else if kogitoRuntimeServices == nil {
		return kdcs, nil
	}

	deploymentHandler := infrastructure.NewDeploymentHandler(k.client, k.log)
	deps, err := deploymentHandler.FetchDeploymentList(namespace)
	if err != nil {
		return nil, err
	}
	k.log.Debug("Looking for Deployments owned by KogitoRuntime")
	for _, dep := range deps.Items {
		for _, owner := range dep.OwnerReferences {
			for _, app := range kogitoRuntimeServices.GetItems() {
				if owner.UID == app.GetUID() {
					kdcs = append(kdcs, dep)
					break
				}
			}
		}
	}
	return kdcs, nil
}
