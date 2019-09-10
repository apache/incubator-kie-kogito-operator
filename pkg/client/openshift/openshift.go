package openshift

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
)

var log = logger.GetLogger("openshift_client")

// ImageStream will call ImageStream OpenShift API
func ImageStream() ImageStreamInterface {
	return newImageStream(&client.Client{})
}

// ImageStreamC will call ImageStream OpenShift API with a given client
func ImageStreamC(c *client.Client) ImageStreamInterface {
	return newImageStream(c)
}

// BuildConfig will call BuildConfig OpenShift API
func BuildConfig() BuildConfigInterface {
	return newBuildConfig(&client.Client{})
}

// BuildConfigC will call BuildConfig OpenShift API with a given client
func BuildConfigC(c *client.Client) BuildConfigInterface {
	return newBuildConfig(c)
}

// Route will call Route OpenShift API
func Route() RouteInterface {
	return newRoute(&client.Client{})
}

// RouteC will call Route OpenShift API with the given client
func RouteC(c *client.Client) RouteInterface {
	return newRoute(c)
}
