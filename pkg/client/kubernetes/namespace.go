package kubernetes

import (
	"context"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// NamespaceInterface has functions that interacts with namespace object in the Kubernetes cluster
type NamespaceInterface interface {
	Fetch(name string) (*corev1.Namespace, error)
	Create(name string) (*corev1.Namespace, error)
	CreateIfNotExists(name string) (*corev1.Namespace, error)
}

type namespace struct {
	client *client.Client
}

func newNamespace(c *client.Client) *namespace {
	if c == nil {
		c = &client.Client{}
	}
	c.ControlCli = client.MustEnsureClient(c)
	return &namespace{
		client: c,
	}
}

func (n *namespace) Fetch(name string) (*corev1.Namespace, error) {
	log.Debugf("About to fetch namespace %s from cluster", name)
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
	if err := n.client.ControlCli.Get(context.TODO(), types.NamespacedName{Name: name}, ns); err != nil && errors.IsNotFound(err) {
		log.Debugf("Namespace %s not found", name)
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return ns, nil
}

func (n *namespace) Create(name string) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
	if err := n.client.ControlCli.Create(context.TODO(), ns); err != nil {
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
