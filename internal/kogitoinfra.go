package internal

import (
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"k8s.io/apimachinery/pkg/types"
)

type kogitoInfraHandler struct {
	client *client.Client
	log    logger.Logger
}

// NewKogitoInfraHandler ...
func NewKogitoInfraHandler(client *client.Client, log logger.Logger) api.KogitoInfraHandler {
	return &kogitoInfraHandler{
		client: client,
		log:    log,
	}
}

// FetchKogitoInfraInstance loads a given infra instance by name and namespace.
// If the KogitoInfra resource is not present, nil will return.
func (k *kogitoInfraHandler) FetchKogitoInfraInstance(key types.NamespacedName) (api.KogitoInfraInterface, error) {
	k.log.Debug("going to fetch deployed kogito infra instance")
	instance := &v1beta1.KogitoInfra{}
	if exists, resultErr := kubernetes.ResourceC(k.client).FetchWithKey(key, instance); resultErr != nil {
		k.log.Error(resultErr, "Error occurs while fetching deployed kogito infra instance")
		return nil, resultErr
	} else if !exists {
		return nil, nil
	} else {
		k.log.Debug("Successfully fetch deployed kogito infra reference")
		return instance, nil
	}
}
