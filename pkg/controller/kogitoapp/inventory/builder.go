package inventory

import (
	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/definitions"
	cliimgv1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BuilderContext is the context for building KogitoApp resources
type BuilderContext struct {
	//KogitoApp is the cached instance of the created CR
	KogitoApp *v1alpha1.KogitoApp
	//Client is the Kubernetes client API
	Client client.Client
	//ImageClient is the OpenShift exclusive Image(tags/stream) API
	ImageClient cliimgv1.ImageV1Interface
}

type builderChain struct {
	Context   *BuilderContext
	Inventory *KogitoAppInventory
	Error     error
}

func (c *builderChain) AndThenCreate(f func(*builderChain) *builderChain) *builderChain {
	if c.Error == nil {
		return f(c)
	}
	// break the chain
	return c
}

// CreateResources will fetch for every resource in KogitoAppInventory and will create it in the cluster if not exists
func CreateResources(context *BuilderContext) (inv *KogitoAppInventory, err error) {
	chain := &builderChain{
		Inventory: &KogitoAppInventory{},
		Context:   context,
	}
	chain.AndThenCreate(serviceAccountBuilder).
		AndThenCreate(roleBindingBuilder).
		AndThenCreate(buildConfigS2IBuilder).
		AndThenCreate(buildConfigServiceBuilder).
		AndThenCreate(imageStreamBuilder).
		AndThenCreate(deploymentConfigBuilder).
		AndThenCreate(serviceBuilder).
		AndThenCreate(routeBuilder)
	return chain.Inventory, chain.Error
}

func serviceAccountBuilder(chain *builderChain) *builderChain {
	sa := definitions.NewServiceAccount(chain.Context.KogitoApp)
	chain.Inventory.ServiceAccountStatus.IsNew, chain.Error = CreateResourceIfNotExists(chain.Context.Client, &sa)
	chain.Inventory.ServiceAccount = &sa
	return chain
}

func roleBindingBuilder(chain *builderChain) *builderChain {
	rb := definitions.NewRoleBinding(chain.Context.KogitoApp, chain.Inventory.ServiceAccount)
	chain.Inventory.RoleBindingStatus.IsNew, chain.Error = CreateResourceIfNotExists(chain.Context.Client, &rb)
	chain.Inventory.RoleBinding = &rb
	return chain
}

func buildConfigS2IBuilder(chain *builderChain) *builderChain {
	bc, err := definitions.NewBuildConfigS2I(chain.Context.KogitoApp)
	if err != nil {
		chain.Error = err
		return chain
	}
	chain.Inventory.BuildConfigS2IStatus.IsNew, chain.Error = CreateResourceIfNotExists(chain.Context.Client, &bc)
	chain.Inventory.BuildConfigS2I = &bc
	return chain
}

func buildConfigServiceBuilder(chain *builderChain) *builderChain {
	bc, err := definitions.NewBuildConfigService(
		chain.Context.KogitoApp,
		chain.Inventory.BuildConfigS2I,
	)
	if err != nil {
		chain.Error = err
		return chain
	}
	chain.Inventory.BuildConfigServiceStatus.IsNew, chain.Error = CreateResourceIfNotExists(chain.Context.Client, &bc)
	chain.Inventory.BuildConfigService = &bc
	return chain
}

func imageStreamBuilder(chain *builderChain) *builderChain {
	_, chain.Error = CreateImageStreamTagIfNotExists(
		chain.Context.ImageClient,
		definitions.NewImageStreamTag(chain.Context.KogitoApp, chain.Inventory.BuildConfigS2I.Name),
	)
	if chain.Error != nil {
		return chain
	}
	_, chain.Error = CreateImageStreamTagIfNotExists(
		chain.Context.ImageClient,
		definitions.NewImageStreamTag(chain.Context.KogitoApp, chain.Inventory.BuildConfigService.Name),
	)
	if chain.Error != nil {
		return chain
	}
	return chain
}

func deploymentConfigBuilder(chain *builderChain) *builderChain {
	bcNamespacedName := types.NamespacedName{
		Namespace: chain.Inventory.BuildConfigService.Namespace,
		Name:      chain.Inventory.BuildConfigService.Name,
	}
	chain.Inventory.DeploymentConfig = nil

	if dockerImage, err := FetchDockerImage(chain.Context.ImageClient, bcNamespacedName); err != nil {
		chain.Error = err
	} else if dockerImage != nil {
		dc, err := definitions.NewDeploymentConfig(
			chain.Context.KogitoApp,
			chain.Inventory.BuildConfigService,
			chain.Inventory.ServiceAccount,
			dockerImage,
		)
		if err != nil {
			chain.Error = err
			return chain
		}
		chain.Inventory.DeploymentConfigStatus.IsNew, chain.Error = CreateResourceIfNotExists(chain.Context.Client, dc)
		chain.Inventory.DeploymentConfig = dc
	} else {
		// TODO: make sure that the build is running
		log.Warnf("Couldn't find an image with name '%s' in the namespace '%s'. The DeploymentConfig will be created once the build is done.", bcNamespacedName.Name, bcNamespacedName.Namespace)
	}
	return chain
}

func serviceBuilder(chain *builderChain) *builderChain {
	// Service depends on the DC
	if chain.Inventory.DeploymentConfig != nil {
		svc := definitions.NewService(chain.Context.KogitoApp, chain.Inventory.DeploymentConfig)
		if svc != nil {
			chain.Inventory.ServiceStatus.IsNew, chain.Error = CreateResourceIfNotExists(chain.Context.Client, svc)
			chain.Inventory.Service = svc
		}
	}
	return chain
}

func routeBuilder(chain *builderChain) *builderChain {
	// we only create a router if we already have a service
	if chain.Inventory.Service != nil {
		route, err := definitions.NewRoute(chain.Context.KogitoApp, chain.Inventory.Service)
		if err != nil {
			chain.Error = err
			return chain
		}
		chain.Inventory.RouteStatus.IsNew, chain.Error = CreateResourceIfNotExists(chain.Context.Client, route)
		chain.Inventory.Route = route
	}
	return chain
}
