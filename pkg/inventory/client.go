package inventory

import (
	"path/filepath"

	"github.com/kiegroup/kogito-cloud-operator/pkg/util"

	"github.com/kiegroup/kogito-cloud-operator/pkg/log"
	buildv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	controllercli "sigs.k8s.io/controller-runtime/pkg/client"
)

// Client wraps clients functions from controller-runtime and OpenShift cli for generic API calls to the cluster
type Client struct {
	// Cli is a reference for the controller-runtime client, normally built by a Manager inside the controller context.
	// If nil, a new client will be created based on a kube configuration file in the host machine.
	Cli      controllercli.Client
	BuildCli buildv1.BuildV1Client
	ImageCli imagev1.ImageV1Client
}

func (c *Client) ensureClient() error {
	if c.Cli == nil {
		config, err := c.buildKubeConnectionConfig()
		if err != nil {
			return err
		}
		c.Cli, err = controllercli.New(config, controllercli.Options{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) buildKubeConnectionConfig() (*restclient.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", *c.getKubeConfigFile())
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (c *Client) getKubeConfigFile() *string {
	var kubeconfig string
	log.Debug("Trying to get kube config file from HOME directory")
	if home := util.GetHomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		log.Warn("Can't read HOME environment variable")
		kubeconfig = filepath.Join(".kube", "config")
	}
	log.Debug("Kube config file read from: ", kubeconfig)
	return &kubeconfig
}
