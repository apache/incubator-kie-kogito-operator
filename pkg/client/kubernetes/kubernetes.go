package kubernetes

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
)

var log = logger.GetLogger("kubernetes_client")

// Namespace will fetch the inner Kubernetes API with a default client
func Namespace() NamespaceInterface {
	return newNamespace(&client.Client{})
}

// NamespaceC will use a defined client to fetch the Kubernetes API
func NamespaceC(c *client.Client) NamespaceInterface {
	return newNamespace(c)
}

// Resource will fetch the inner API for any Kubernetes resource with a default client
func Resource() ResourceInterface {
	return newResource(&client.Client{})
}

// ResourceC will use a defined client to fetch the Kubernetes API
func ResourceC(c *client.Client) ResourceInterface {
	return newResource(c)
}
