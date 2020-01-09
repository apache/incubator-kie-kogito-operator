// Copyright 2019 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"fmt"
	appsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	olmapiv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1"
	olmapiv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"strings"

	operatormkt "github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"
	coreappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	controllercli "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"

	ocappsv1 "github.com/openshift/api/apps/v1"
	ocroutev1 "github.com/openshift/api/route/v1"
	buildv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"

	monclientv1 "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
)

const (
	envVarKubeConfig = "KUBECONFIG"
)

var (
	log                   = logger.GetLogger("client_api")
	defaultKubeConfigPath = filepath.Join(".kube", "config")
)

// Client wraps clients functions from controller-runtime, Kube and OpenShift cli for generic API calls to the cluster
type Client struct {
	// ControlCli is a reference for the controller-runtime client, normally built by a Manager inside the controller context.
	ControlCli    controllercli.Client
	BuildCli      buildv1.BuildV1Interface
	ImageCli      imagev1.ImageV1Interface
	Discovery     discovery.DiscoveryInterface
	PrometheusCli monclientv1.MonitoringV1Interface
	DeploymentCli appsv1.AppsV1Interface
}

// NewForConsole will create a brand new client using the local machine
func NewForConsole() *Client {
	client, err := NewClientBuilder().WithDiscoveryClient().Build()
	if err != nil {
		panic(err)
	}
	return client
}

// WrapperForManager creates a wrapper around the manager client, useful for small operations that requires the client interface.
// only creates the ControlCli reference
func WrapperForManager(mgr manager.Manager) *Client {
	return &Client{
		ControlCli: mgr.GetClient(),
	}
}

// NewForController creates a new client based on the rest config and the controller client created by Operator SDK
// Panic if something goes wrong
func NewForController(config *restclient.Config, client controllercli.Client) *Client {
	newClient, err := NewClientBuilder().WithAllClients().UseConfig(config).UseControllerClient(client).Build()
	if err != nil {
		panic(err)
	}
	return newClient
}

// IsOpenshift detects if the application is running on OpenShift or not
func (c *Client) IsOpenshift() bool {
	return c.HasServerGroup("openshift.io")
}

// HasServerGroup detects if the given api group is supported by the server
func (c *Client) HasServerGroup(groupName string) bool {
	if c.Discovery != nil {
		groups, err := c.Discovery.ServerGroups()
		if err != nil {
			log.Warnf("Impossible to get server groups using discovery API: %s", err)
			return false
		}
		for _, group := range groups.Groups {
			if strings.Contains(group.Name, groupName) {
				return true
			}
		}
		return false
	}
	log.Warnf("Tried to discover the platform, but no discovery API is available")
	return false
}

// MustEnsureClient will try to read the kube.yaml file from the host and connect to the cluster, if the Client or the Core Client is null.
// Will panic if the connection won't be possible
func MustEnsureClient(c *Client) controllercli.Client {
	if c.ControlCli == nil {
		// fallback to the KubeClient
		var err error
		if c.ControlCli, err = ensureKubeClient(); err != nil {
			panic(fmt.Sprintf("Error while trying to create a new kubernetes client: %s", err))
		}
	}

	return c.ControlCli
}

func ensureKubeClient() (controllercli.Client, error) {
	log.Debugf("Veryfing kube core client dependencies")
	config, err := buildKubeConnectionConfig()
	if err != nil {
		return nil, err
	}
	return newKubeClient(config)
}

func newKubeClient(config *restclient.Config) (controllercli.Client, error) {
	log.Debugf("Creating a new core client for kube connection")
	controlCli, err := controllercli.New(config, newControllerCliOptions())
	if err != nil {
		return nil, err
	}
	return controlCli, nil
}

func buildKubeConnectionConfig() (*restclient.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", getKubeConfigFile())
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getKubeConfigFile() string {
	kubeconfig := util.GetOSEnv(envVarKubeConfig, "")
	if len(kubeconfig) > 0 {
		log.Debugf("Kube config file read from %s environment variable: %s", envVarKubeConfig, kubeconfig)
		return kubeconfig
	}
	log.Debug("Trying to get kube config file from HOME directory")
	if home := util.GetHomeDir(); home != "" {
		kubeconfig = filepath.Join(home, defaultKubeConfigPath)
	} else {
		log.Warn("Can't read HOME environment variable")
		kubeconfig = defaultKubeConfigPath
	}
	log.Debug("Kube config file read from: ", kubeconfig)
	return kubeconfig
}

//restScope implementation
type restScope struct {
	name apimeta.RESTScopeName
}

func (r *restScope) Name() apimeta.RESTScopeName {
	return r.name
}

// newControllerCliOptions creates the mapper and schema options for the inner fallback cli. If set to defaults, the Controller Cli will try
// to discover the mapper by itself by querying the API, which can take too much time. Here we're setting this mapper manually.
// So it's need to keep adding them or find some kind of auto register in the kube api/apimachinery
func newControllerCliOptions() controllercli.Options {
	options := controllercli.Options{}

	mapper := apimeta.NewDefaultRESTMapper([]schema.GroupVersion{})
	mapper.Add(corev1.SchemeGroupVersion.WithKind(meta.KindNamespace.Name), &restScope{name: apimeta.RESTScopeNameRoot})
	mapper.Add(corev1.SchemeGroupVersion.WithKind(meta.KindServiceAccount.Name), &restScope{name: apimeta.RESTScopeNameNamespace})
	mapper.Add(apiextensionsv1beta1.SchemeGroupVersion.WithKind(meta.KindCRD.Name), &restScope{name: apimeta.RESTScopeNameRoot})
	mapper.Add(v1alpha1.SchemeGroupVersion.WithKind(meta.KindKogitoApp.Name), &restScope{name: apimeta.RESTScopeNameNamespace})
	mapper.Add(v1alpha1.SchemeGroupVersion.WithKind(meta.KindKogitoJobsService.Name), &restScope{name: apimeta.RESTScopeNameNamespace})
	mapper.Add(coreappsv1.SchemeGroupVersion.WithKind(meta.KindDeployment.Name), &restScope{name: apimeta.RESTScopeNameNamespace})
	mapper.Add(rbac.SchemeGroupVersion.WithKind(meta.KindRole.Name), &restScope{name: apimeta.RESTScopeNameNamespace})
	mapper.Add(rbac.SchemeGroupVersion.WithKind(meta.KindRoleBinding.Name), &restScope{name: apimeta.RESTScopeNameNamespace})
	mapper.Add(operatormkt.SchemeGroupVersion.WithKind(meta.KindOperatorSource.Name), &restScope{name: apimeta.RESTScopeNameNamespace})
	mapper.Add(ocappsv1.SchemeGroupVersion.WithKind(meta.KindDeploymentConfig.Name), &restScope{name: apimeta.RESTScopeNameNamespace})
	mapper.Add(ocroutev1.SchemeGroupVersion.WithKind(meta.KindRoute.Name), &restScope{name: apimeta.RESTScopeNameNamespace})
	mapper.Add(olmapiv1.SchemeGroupVersion.WithKind(meta.KindOperatorGroup.Name), &restScope{name: apimeta.RESTScopeNameNamespace})
	mapper.Add(olmapiv1alpha1.SchemeGroupVersion.WithKind(meta.KindSubscription.Name), &restScope{name: apimeta.RESTScopeNameNamespace})

	// the kube client is having problems with plural: kogitodataindexs :(
	mapper.AddSpecific(v1alpha1.SchemeGroupVersion.WithKind(meta.KindKogitoDataIndex.Name),
		schema.GroupVersionResource{
			Group:    meta.KindKogitoDataIndex.GroupVersion.Group,
			Version:  meta.KindKogitoDataIndex.GroupVersion.Version,
			Resource: "kogitodataindices",
		},
		schema.GroupVersionResource{
			Group:    meta.KindKogitoDataIndex.GroupVersion.Group,
			Version:  meta.KindKogitoDataIndex.GroupVersion.Version,
			Resource: "kogitodataindex",
		},
		&restScope{name: apimeta.RESTScopeNameNamespace})

	options.Scheme = meta.GetRegisteredSchema()
	options.Mapper = mapper
	return options
}
