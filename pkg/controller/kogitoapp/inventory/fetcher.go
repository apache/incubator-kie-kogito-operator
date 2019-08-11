package inventory

import (
	"context"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListResourceWithNamespace fetches and binds a list resource from the Kubernetes cluster with the defined namespace.
func ListResourceWithNamespace(cli client.Client, ns string, list runtime.Object) error {
	err := cli.List(context.TODO(), &client.ListOptions{Namespace: ns}, list)
	if err != nil {
		log.Warn("Failed to list resource. ", err)
		return err
	}
	return nil
}

// FetchResource fetches and binds a resource with given name and namespace from the Kubernetes cluster. If not exists, returns false.
func FetchResource(cli client.Client, resource v1alpha1.OpenShiftObject) (bool, error) {
	return FetchResourceWithKey(cli, types.NamespacedName{Name: resource.GetName(), Namespace: resource.GetNamespace()}, resource)
}

// FetchResourceWithKey fetches and binds a resource from the Kubernetes cluster with the defined key. If not exists, returns false.
func FetchResourceWithKey(cli client.Client, key client.ObjectKey, resource v1alpha1.OpenShiftObject) (bool, error) {
	err := cli.Get(context.TODO(), key, resource)
	if err != nil && errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

// CreateResourceIfNotExists creates a Kubernetes/OpenShift resource if not exists in the cluster.
// Returns true if the resource has been created, false otherwise
func CreateResourceIfNotExists(cli client.Client, resource v1alpha1.OpenShiftObject) (bool, error) {
	log := log.With("kind", resource.GetObjectKind().GroupVersionKind().Kind, "name", resource.GetName(), "namespace", resource.GetNamespace())

	if exists, err := FetchResource(cli, resource); err == nil && !exists {
		// Define a new Object
		log.Info("Creating")
		err = cli.Create(context.TODO(), resource)
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
