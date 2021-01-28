package kogitoservice

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/core/manager"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// FinalizerHandler ...
type FinalizerHandler interface {
	AddFinalizer(instance api.KogitoService) error
	HandleFinalization(instance api.KogitoService) error
}

type finalizerHandler struct {
	client       *client.Client
	log          logger.Logger
	scheme       *runtime.Scheme
	infraHandler api.KogitoInfraHandler
}

// NewFinalizerHandler ...
func NewFinalizerHandler(client *client.Client, log logger.Logger, scheme *runtime.Scheme, infraHandler api.KogitoInfraHandler) FinalizerHandler {
	return &finalizerHandler{
		client:       client,
		log:          log,
		scheme:       scheme,
		infraHandler: infraHandler,
	}
}

// AddFinalizer add finalizer to provide KogitoService instance
func (f *finalizerHandler) AddFinalizer(instance api.KogitoService) error {
	if len(instance.GetFinalizers()) < 1 && instance.GetDeletionTimestamp() == nil {
		f.log.Debug("Adding Finalizer for the KogitoService")
		instance.SetFinalizers([]string{"delete.kogitoInfra.ownership.finalizer"})

		// Update CR
		if err := kubernetes.ResourceC(f.client).Update(instance); err != nil {
			f.log.Error(err, "Failed to update finalizer in KogitoService")
			return err
		}
		f.log.Debug("Successfully added finalizer into KogitoService instance", "instance", instance.GetName())
	}
	return nil
}

// HandleFinalization remove owner reference of provided Kogito service from KogitoInfra instances and remove finalizer from KogitoService
func (f *finalizerHandler) HandleFinalization(instance api.KogitoService) error {
	// Remove KogitoSupportingService ownership from referred KogitoInfra instances
	infraManager := manager.NewKogitoInfraManager(f.client, f.log, f.scheme, f.infraHandler)
	for _, kogitoInfra := range instance.GetSpec().GetInfra() {
		if err := infraManager.RemoveKogitoInfraOwnership(types.NamespacedName{Name: kogitoInfra, Namespace: instance.GetNamespace()}, instance); err != nil {
			return err
		}
	}
	// Update finalizer to allow delete CR
	f.log.Debug("Removing finalizer from KogitoService")
	instance.SetFinalizers(nil)
	if err := kubernetes.ResourceC(f.client).Update(instance); err != nil {
		f.log.Error(err, "Error occurs while removing finalizer from KogitoService")
		return err
	}
	f.log.Debug("Successfully removed finalizer from KogitoService")
	return nil
}
