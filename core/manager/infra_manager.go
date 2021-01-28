package manager

import (
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/framework"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// KogitoInfraManager ...
type KogitoInfraManager interface {
	MustFetchKogitoInfraInstance(key types.NamespacedName) (api.KogitoInfraInterface, error)
	TakeKogitoInfraOwnership(key types.NamespacedName, owner resource.KubernetesResource) error
	RemoveKogitoInfraOwnership(key types.NamespacedName, owner resource.KubernetesResource) error
	IsKogitoInfraReady(key types.NamespacedName) (bool, error)
	GetKogitoInfraConditionReason(key types.NamespacedName) (api.KogitoInfraConditionReason, error)
}

type kogitoInfraManager struct {
	client       *client.Client
	log          logger.Logger
	scheme       *runtime.Scheme
	infraHandler api.KogitoInfraHandler
}

// NewKogitoInfraManager ...
func NewKogitoInfraManager(client *client.Client, log logger.Logger, scheme *runtime.Scheme, infraHandler api.KogitoInfraHandler) KogitoInfraManager {
	return &kogitoInfraManager{
		client:       client,
		log:          log,
		scheme:       scheme,
		infraHandler: infraHandler,
	}
}

// MustFetchKogitoInfraInstance loads a given infra instance by name and namespace.
// If the KogitoInfra resource is not present, an error is raised.
func (k *kogitoInfraManager) MustFetchKogitoInfraInstance(key types.NamespacedName) (api.KogitoInfraInterface, error) {
	k.log.Debug("going to must fetch deployed kogito infra instance")
	if instance, resultErr := k.infraHandler.FetchKogitoInfraInstance(key); resultErr != nil {
		k.log.Error(resultErr, "Error occurs while fetching deployed kogito infra instance")
		return nil, resultErr
	} else if instance == nil {
		return nil, fmt.Errorf("kogito Infra resource with name %s not found in namespace %s", key.Name, key.Namespace)
	} else {
		k.log.Debug("Successfully fetch deployed kogito infra reference")
		return instance, nil
	}
}

func (k *kogitoInfraManager) TakeKogitoInfraOwnership(key types.NamespacedName, owner resource.KubernetesResource) (err error) {
	kogitoInfra, err := k.MustFetchKogitoInfraInstance(key)
	if err != nil {
		return
	}
	if framework.IsOwner(kogitoInfra, owner) {
		return
	}
	if err = framework.AddOwnerReference(owner, k.scheme, kogitoInfra); err != nil {
		return
	}
	if err = kubernetes.ResourceC(k.client).Update(kogitoInfra); err != nil {
		return
	}
	return
}

func (k *kogitoInfraManager) RemoveKogitoInfraOwnership(key types.NamespacedName, owner resource.KubernetesResource) (err error) {
	k.log.Info("Removing kogito infra ownership", "infra name", key.Name, "owner", owner.GetName())
	kogitoInfra, err := k.MustFetchKogitoInfraInstance(key)
	if err != nil {
		return
	}
	framework.RemoveOwnerReference(owner, kogitoInfra)
	if err = kubernetes.ResourceC(k.client).Update(kogitoInfra); err != nil {
		return err
	}
	k.log.Debug("Successfully removed KogitoInfra ownership", "infra name", kogitoInfra.GetName(), "owner", owner.GetName())
	return
}

func (k *kogitoInfraManager) IsKogitoInfraReady(key types.NamespacedName) (bool, error) {
	infra, err := k.MustFetchKogitoInfraInstance(key)
	if err != nil {
		return false, err
	}
	if infra.GetStatus().GetCondition().Type == api.FailureInfraConditionType {
		return false, nil
	}
	return true, nil
}

func (k *kogitoInfraManager) GetKogitoInfraConditionReason(key types.NamespacedName) (api.KogitoInfraConditionReason, error) {
	infra, err := k.MustFetchKogitoInfraInstance(key)
	if err != nil {
		return "", err
	}

	return infra.GetStatus().GetCondition().Reason, nil
}
