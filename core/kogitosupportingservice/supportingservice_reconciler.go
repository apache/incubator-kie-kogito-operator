package kogitosupportingservice

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	"time"
)

// Reconciler Interface to represent type of kogito supporting service resources like JobsService & MgmtConcole
type Reconciler interface {
	Reconcile() (reconcileAfter time.Duration, resultErr error)
}

type targetContext struct {
	client                   *client.Client
	instance                 api.KogitoSupportingServiceInterface
	scheme                   *runtime.Scheme
	log                      logger.Logger
	infraHandler             api.KogitoInfraHandler
	supportingServiceHandler api.KogitoSupportingServiceHandler
	runtimeHandler           api.KogitoRuntimeHandler
}

// ReconcilerHandler ...
type ReconcilerHandler interface {
	GetSupportingServiceReconciler(instance api.KogitoSupportingServiceInterface) Reconciler
}

type reconcilerHandler struct {
	client                   *client.Client
	log                      logger.Logger
	scheme                   *runtime.Scheme
	infraHandler             api.KogitoInfraHandler
	supportingServiceHandler api.KogitoSupportingServiceHandler
	runtimeHandler           api.KogitoRuntimeHandler
}

// NewReconcilerHandler ...
func NewReconcilerHandler(cli *client.Client, log logger.Logger, scheme *runtime.Scheme, infraHandler api.KogitoInfraHandler, supportingServiceHandler api.KogitoSupportingServiceHandler, runtimeHandler api.KogitoRuntimeHandler) ReconcilerHandler {
	return &reconcilerHandler{
		client:                   cli,
		log:                      log,
		scheme:                   scheme,
		infraHandler:             infraHandler,
		supportingServiceHandler: supportingServiceHandler,
		runtimeHandler:           runtimeHandler,
	}
}

// getKogitoInfraReconciler identify and return request kogito infra reconciliation logic on bases of information provided in kogitoInfra value
func (k *reconcilerHandler) GetSupportingServiceReconciler(instance api.KogitoSupportingServiceInterface) Reconciler {
	k.log.Debug("going to fetch related kogito supporting service resource")
	context := targetContext{
		client:                   k.client,
		instance:                 instance,
		scheme:                   k.scheme,
		log:                      k.log,
		infraHandler:             k.infraHandler,
		supportingServiceHandler: k.supportingServiceHandler,
		runtimeHandler:           k.runtimeHandler,
	}
	return getSupportedResources(context)[instance.GetSupportingServiceSpec().GetServiceType()]
}

func getSupportedResources(context targetContext) map[api.ServiceType]Reconciler {
	return map[api.ServiceType]Reconciler{
		api.DataIndex:      initDataIndexSupportingServiceResource(context),
		api.Explainability: initExplainabilitySupportingServiceResource(context),
		api.JobsService:    initJobsServiceSupportingServiceResource(context),
		api.MgmtConsole:    initMgmtConsoleSupportingServiceResource(context),
		api.TaskConsole:    initTaskConsoleSupportingServiceResource(context),
		api.TrustyAI:       initTrustyAISupportingServiceResource(context),
		api.TrustyUI:       initTrustyUISupportingServiceResource(context),
	}
}
