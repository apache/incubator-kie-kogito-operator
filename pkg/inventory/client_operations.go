package inventory

import (
	"context"

	"github.com/kiegroup/kogito-cloud-operator/pkg/log"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func (c *Client) fetch(key types.NamespacedName, object runtime.Object) error {
	log.Debugf("About to fetch object '%s' on namespace '%s'", key.Name, key.Namespace)
	if err := c.ensureClient(); err != nil {
		return err
	}
	if err := c.Cli.Get(context.TODO(), key, object); err != nil {
		return err
	}
	return nil
}

func (c *Client) create(object runtime.Object) error {
	log.Debug("About to create the object")
	if err := c.ensureClient(); err != nil {
		return err
	}
	if err := c.Cli.Create(context.TODO(), object); err != nil {
		return err
	}
	log.Debug("Object created")
	return nil
}
