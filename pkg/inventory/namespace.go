package inventory

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// NamespaceInventory is a interface that exposes the namespace object functions.
// Should be exposed by inventory implementations
type NamespaceInventory interface {
	Namespace(c *Client) NamespaceInterface
}

// NamespaceInterface has functions that interacts with namespace object in the Kubernetes cluster
type NamespaceInterface interface {
	Fetch(string) (*corev1.Namespace, error)
	Create(string) (*corev1.Namespace, error)
	CreateIfNotExists(string) (*corev1.Namespace, error)
}

type namespace struct {
	client *Client
}

func newNamespace(c *Client) *namespace {
	return &namespace{
		client: c,
	}
}

func (n *namespace) Fetch(name string) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
	if err := n.client.fetch(types.NamespacedName{Name: name}, ns); err != nil && errors.IsNotFound(err) {
		log.Debugf("Namespace %s not found", name)
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return ns, nil
}

func (n *namespace) Create(name string) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
	if err := n.client.create(ns); err != nil {
		return nil, err
	}
	return ns, nil
}

func (n *namespace) CreateIfNotExists(name string) (*corev1.Namespace, error) {
	if ns, err := n.Fetch(name); err != nil {
		return nil, err
	} else if ns != nil {
		return ns, nil
	}
	ns, err := n.Create(name)
	if err != nil {
		return nil, err
	}
	return ns, nil
}
