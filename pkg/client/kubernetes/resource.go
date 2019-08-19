package kubernetes

import (
	"context"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"

	runtimecli "sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// ResourceInterface has functions that interacts with any resource object in the Kubernetes cluster
type ResourceInterface interface {
	CreateIfNotExists(kubeObject meta.ResourceObject) (bool, error)
	FetchWithKey(key types.NamespacedName, kubeObject meta.ResourceObject) (bool, error)
	Fetch(kubeObject meta.ResourceObject) (bool, error)
	ListWithNamespace(namespace string, list runtime.Object) error
}

type resource struct {
	client *client.Client
}

func newResource(c *client.Client) *resource {
	if c == nil {
		c = &client.Client{}
	}
	c.ControlCli = client.MustEnsureClient(c)
	return &resource{
		client: c,
	}
}

// Fetch fetches and binds a resource with given name and namespace from the Kubernetes cluster. If not exists, returns false.
func (r *resource) Fetch(resource meta.ResourceObject) (bool, error) {
	return r.FetchWithKey(types.NamespacedName{Name: resource.GetName(), Namespace: resource.GetNamespace()}, resource)
}

// FetchWithKey fetches and binds a resource from the Kubernetes cluster with the defined key. If not exists, returns false.
func (r *resource) FetchWithKey(key types.NamespacedName, resource meta.ResourceObject) (bool, error) {
	log.Debugf("About to fetch object '%s' on namespace '%s'", key.Name, key.Namespace)
	err := r.client.ControlCli.Get(context.TODO(), key, resource)
	if err != nil && errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	log.Debugf("Found object (%s) '%s' in the namespace '%s'. Creation time is: %s",
		resource.GetObjectKind().GroupVersionKind().Kind,
		key.Name,
		key.Namespace,
		resource.GetCreationTimestamp())
	return true, nil
}

// CreateIfNotExists will fetch for the object resource in the Kubernetes cluster, if not exists, will create it.
func (r *resource) CreateIfNotExists(resource meta.ResourceObject) (bool, error) {
	log := log.With("kind", resource.GetObjectKind().GroupVersionKind().Kind, "name", resource.GetName(), "namespace", resource.GetNamespace())

	if exists, err := r.Fetch(resource); err == nil && !exists {
		// Define a new Object
		log.Info("Creating")
		err = r.client.ControlCli.Create(context.TODO(), resource)
		if err != nil {
			log.Warn("Failed to create object. ", err)
			return false, err
		}
		return true, nil
	} else if err != nil {
		log.Warn("Failed to fecth object. ", err)
		return false, err
	}

	log.Debug("Skip creating - object already exists")
	return false, nil
}

// ListWithNamespace fetches and binds a list resource from the Kubernetes cluster with the defined namespace.
func (r *resource) ListWithNamespace(ns string, list runtime.Object) error {
	err := r.client.ControlCli.List(context.TODO(), &runtimecli.ListOptions{Namespace: ns}, list)
	if err != nil {
		log.Warn("Failed to list resource. ", err)
		return err
	}
	return nil
}
