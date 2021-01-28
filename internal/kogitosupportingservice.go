package internal

import (
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"k8s.io/apimachinery/pkg/types"
)

type kogitoSupportingServiceHandler struct {
	client *client.Client
	log    logger.Logger
}

// NewKogitoSupportingServiceHandler ...
func NewKogitoSupportingServiceHandler(client *client.Client, log logger.Logger) api.KogitoSupportingServiceHandler {
	return &kogitoSupportingServiceHandler{
		client: client,
		log:    log,
	}
}

// FetchKogitoSupportingService provide kogito supporting service instance
func (k kogitoSupportingServiceHandler) FetchKogitoSupportingService(key types.NamespacedName) (api.KogitoSupportingServiceInterface, error) {
	k.log.Info("going to fetch deployed kogito supporting service")
	instance := &v1beta1.KogitoSupportingService{}
	if exists, resultErr := kubernetes.ResourceC(k.client).FetchWithKey(key, instance); resultErr != nil {
		k.log.Error(resultErr, "Error occurs while fetching deployed kogito supporting service")
		return nil, resultErr
	} else if !exists {
		return nil, nil
	} else {
		k.log.Debug("Successfully fetch deployed kogito supporting reference")
		return instance, nil
	}
}

func (k kogitoSupportingServiceHandler) FetchKogitoSupportingServiceList(namespace string) (api.KogitoSupportingServiceListInterface, error) {
	supportingServiceList := &v1beta1.KogitoSupportingServiceList{}
	if err := kubernetes.ResourceC(k.client).ListWithNamespace(namespace, supportingServiceList); err != nil {
		return nil, err
	}
	return supportingServiceList, nil
}
