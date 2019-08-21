package builder

import (
	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"k8s.io/apimachinery/pkg/types"
)

// Context is the context for building KogitoApp resources
type Context struct {
	//KogitoApp is the cached instance of the created CR
	KogitoApp *v1alpha1.KogitoApp
	Client    *client.Client
}

type builderChain struct {
	Context   *Context
	Inventory *KogitoAppInventory
	Error     error
}

func (c *builderChain) AndBuild(f func(*builderChain) *builderChain) *builderChain {
	if c.Error == nil {
		return f(c)
	}
	// break the chain
	return c
}

// BuildOrFetchObjects will fetch for every resource in KogitoAppInventory and will create it in the cluster if not exists
func BuildOrFetchObjects(context *Context) (inv *KogitoAppInventory, err error) {
	chain := &builderChain{
		Inventory: &KogitoAppInventory{},
		Context:   context,
	}
	chain.AndBuild(serviceAccountBuilder).
		AndBuild(roleBuilder).
		AndBuild(roleBindingBuilder).
		AndBuild(buildConfigS2IBuilder).
		AndBuild(buildConfigServiceBuilder).
		AndBuild(imageStreamBuilder).
		AndBuild(deploymentConfigBuilder).
		AndBuild(serviceBuilder).
		AndBuild(routeBuilder)
	return chain.Inventory, chain.Error
}

func serviceAccountBuilder(chain *builderChain) *builderChain {
	sa := NewServiceAccount(chain.Context.KogitoApp)
	chain.Inventory.ServiceAccountStatus.IsNew, chain.Error =
		kubernetes.ResourceC(chain.Context.Client).CreateIfNotExists(&sa)
	chain.Inventory.ServiceAccount = &sa
	return chain
}

func roleBuilder(chain *builderChain) *builderChain {
	role := NewRole(chain.Context.KogitoApp)
	chain.Inventory.RoleStatus.IsNew, chain.Error =
		kubernetes.ResourceC(chain.Context.Client).CreateIfNotExists(&role)
	chain.Inventory.Role = &role
	return chain
}

func roleBindingBuilder(chain *builderChain) *builderChain {
	rb := NewRoleBinding(chain.Context.KogitoApp, chain.Inventory.ServiceAccount, chain.Inventory.Role)
	chain.Inventory.RoleBindingStatus.IsNew, chain.Error =
		kubernetes.ResourceC(chain.Context.Client).CreateIfNotExists(&rb)
	chain.Inventory.RoleBinding = &rb
	return chain
}

func buildConfigS2IBuilder(chain *builderChain) *builderChain {
	bc, err := NewBuildConfigS2I(chain.Context.KogitoApp)
	if err != nil {
		chain.Error = err
		return chain
	}
	chain.Inventory.BuildConfigS2IStatus.IsNew, chain.Error =
		kubernetes.ResourceC(chain.Context.Client).CreateIfNotExists(&bc)
	chain.Inventory.BuildConfigS2I = &bc
	return chain
}

func buildConfigServiceBuilder(chain *builderChain) *builderChain {
	bc, err := NewBuildConfigService(
		chain.Context.KogitoApp,
		chain.Inventory.BuildConfigS2I,
	)
	if err != nil {
		chain.Error = err
		return chain
	}
	chain.Inventory.BuildConfigServiceStatus.IsNew, chain.Error =
		kubernetes.ResourceC(chain.Context.Client).CreateIfNotExists(&bc)
	chain.Inventory.BuildConfigService = &bc
	return chain
}

func imageStreamBuilder(chain *builderChain) *builderChain {
	_, chain.Error =
		openshift.ImageStreamC(chain.Context.Client).CreateTagIfNotExists(
			NewImageStreamTag(chain.Context.KogitoApp, chain.Inventory.BuildConfigS2I.Name),
		)
	if chain.Error != nil {
		return chain
	}
	_, chain.Error =
		openshift.ImageStreamC(chain.Context.Client).CreateTagIfNotExists(
			NewImageStreamTag(chain.Context.KogitoApp, chain.Inventory.BuildConfigService.Name),
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

	if dockerImage, err := openshift.ImageStreamC(chain.Context.Client).FetchDockerImage(bcNamespacedName); err != nil {
		chain.Error = err
	} else if dockerImage != nil {
		dc, err := NewDeploymentConfig(
			chain.Context.KogitoApp,
			chain.Inventory.BuildConfigService,
			chain.Inventory.ServiceAccount,
			dockerImage,
		)
		if err != nil {
			chain.Error = err
			return chain
		}
		chain.Inventory.DeploymentConfigStatus.IsNew, chain.Error =
			kubernetes.ResourceC(chain.Context.Client).CreateIfNotExists(dc)
		chain.Inventory.DeploymentConfig = dc
	} else {
		log.Warnf("Couldn't find an image with name '%s' in the namespace '%s'. The DeploymentConfig will be created once the build is done.", bcNamespacedName.Name, bcNamespacedName.Namespace)
	}
	return chain
}

func serviceBuilder(chain *builderChain) *builderChain {
	// Service depends on the DC
	if chain.Inventory.DeploymentConfig != nil {
		svc := NewService(chain.Context.KogitoApp, chain.Inventory.DeploymentConfig)
		if svc != nil {
			chain.Inventory.ServiceStatus.IsNew, chain.Error =
				kubernetes.ResourceC(chain.Context.Client).CreateIfNotExists(svc)
			chain.Inventory.Service = svc
		}
	}
	return chain
}

func routeBuilder(chain *builderChain) *builderChain {
	// we only create a router if we already have a service
	if chain.Inventory.Service != nil {
		route, err := NewRoute(chain.Context.KogitoApp, chain.Inventory.Service)
		if err != nil {
			chain.Error = err
			return chain
		}
		chain.Inventory.RouteStatus.IsNew, chain.Error =
			kubernetes.ResourceC(chain.Context.Client).CreateIfNotExists(route)
		chain.Inventory.Route = route
	}
	return chain
}
