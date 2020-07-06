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

	"os"
	"path/filepath"
	"strings"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"

	appsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	controllercli "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	buildv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"

	monclientv1 "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
)

var (
	log = logger.GetLogger("client_api")
)

const (
	// OpenShiftGroupName name included in OpenShift APIs
	OpenShiftGroupName = "openshift.io"
)

// Client wraps clients functions from controller-runtime, Kube and OpenShift cli for generic API calls to the cluster
type Client struct {
	// ControlCli is a reference for the controller-runtime client, normally built by a Manager inside the controller context.
	ControlCli             controllercli.Client
	BuildCli               buildv1.BuildV1Interface
	ImageCli               imagev1.ImageV1Interface
	Discovery              discovery.DiscoveryInterface
	PrometheusCli          monclientv1.MonitoringV1Interface
	DeploymentCli          appsv1.AppsV1Interface
	KubernetesExtensionCli kubernetes.Interface
}

// NewForConsole will create a brand new client using the local machine
func NewForConsole() *Client {
	client, err := NewClientBuilder().WithDiscoveryClient().Build()
	if err != nil {
		panic(err)
	}
	return client
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
	return c.HasServerGroup(OpenShiftGroupName)
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
	config, err := clientcmd.BuildConfigFromFlags("", GetKubeConfigFile())
	if err != nil {
		return nil, err
	}
	return config, nil
}

// GetKubeConfigFile gets the .kubeconfig file.
// Never returns an empty string, fallback to default path if not present in the known locations
func GetKubeConfigFile() string {
	filename := clientcmd.NewDefaultPathOptions().GetDefaultFilename()
	// make sure the path to the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		dirName := filepath.Dir(filename)
		if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
			panic(fmt.Errorf("Error while trying to create kube config directories %s: %s ", filename, err))
		}
	} else if err != nil {
		panic(fmt.Errorf("Error while trying to access the kube config file %s: %s ", filename, err))
	}
	if file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0600); err != nil {
		panic(fmt.Errorf("Error while trying to access the kube config file %s: %s ", filename, err))
	} else {
		defer func() {
			if file != nil {
				if err := file.Close(); err != nil {
					panic(fmt.Errorf("Error closing kube config file %s: %s ", filename, err))
				}
			}
		}()
		if fileInfo, err := file.Stat(); err != nil {
			panic(fmt.Errorf("Error while trying to access the kube config file %s: %s ", filename, err))
		} else if fileInfo.Size() == 0 {
			log.Warnf("Kubernetes local configuration '%s' is empty.", filename)
			log.Warn("Make sure to login to your cluster with oc/kubectl before using this tool")
		}
	}

	return filename
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
	builder := meta.GetRegisteredSchemeBuilder()
	gvks, err := k8sutil.GetGVKsFromAddToScheme(builder.AddToScheme)
	if err != nil {
		log.Fatalf("Error while creating SchemeBuilder for Kubernetes client: %v", err)
		panic(err)
	}
	mapper := apimeta.NewDefaultRESTMapper([]schema.GroupVersion{})
	for _, gvk := range gvks {
		// we will handle this manually later
		if gvk.GroupVersion() == v1alpha1.SchemeGroupVersion && gvk.Kind == "KogitoDataIndex" {
			continue
		}
		// namespaced resources
		if (gvk.GroupVersion() == corev1.SchemeGroupVersion && gvk.Kind == "Namespace") ||
			(gvk.GroupVersion() == apiextensionsv1beta1.SchemeGroupVersion && gvk.Kind == "CustomResourceDefinition") {
			mapper.Add(gvk, &restScope{name: apimeta.RESTScopeNameRoot})
		} else { // everything else
			mapper.Add(gvk, &restScope{name: apimeta.RESTScopeNameNamespace})
		}
	}
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
