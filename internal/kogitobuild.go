package internal

import (
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"k8s.io/apimachinery/pkg/types"
)

type kogitoBuildHandler struct {
	client *client.Client
	log    logger.Logger
}

// NewKogitoBuildHandler ...
func NewKogitoBuildHandler(client *client.Client, log logger.Logger) api.KogitoBuildHandler {
	return &kogitoBuildHandler{
		client: client,
		log:    log,
	}
}

func (k *kogitoBuildHandler) FetchKogitoBuildInstance(key types.NamespacedName) (api.KogitoBuildInterface, error) {
	instance := &v1beta1.KogitoBuild{}
	if exists, err := kubernetes.ResourceC(k.client).FetchWithKey(key, instance); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}
	return instance, nil
}
