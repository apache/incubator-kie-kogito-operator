package test

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	api2 "github.com/kiegroup/kogito-cloud-operator/core/test/api"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"k8s.io/apimachinery/pkg/types"
)

type fakeKogitoBuildHandler struct {
	client *client.Client
}

// CreateFakeKogitoBuildHandler ...
func CreateFakeKogitoBuildHandler(client *client.Client) api.KogitoBuildHandler {
	return &fakeKogitoBuildHandler{
		client: client,
	}
}

func (k *fakeKogitoBuildHandler) FetchKogitoBuildInstance(key types.NamespacedName) (api.KogitoBuildInterface, error) {
	instance := &api2.KogitoBuildTest{}
	if exists, err := kubernetes.ResourceC(k.client).FetchWithKey(key, instance); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}
	return instance, nil
}
