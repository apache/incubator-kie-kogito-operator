package services

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ServiceEndPoint_String(t *testing.T) {
	serviceEndPoint := infrastructure.ServiceEndpoints{
		HTTPRouteURI: "HTTPRouteURI",
		HTTPRouteEnv: "HTTPRouteEnv",
		WSRouteURI:   "WSRouteURI",
		WSRouteEnv:   "WSRouteEnv",
	}

	assert.Equal(t, serviceEndPoint.HTTPRouteURI, serviceEndPoint.String())
}

func Test_ServiceEndPoint_IsEmpty(t *testing.T) {
	serviceEndPoint := infrastructure.ServiceEndpoints{
		HTTPRouteEnv: "HTTPRouteEnv",
		WSRouteEnv:   "WSRouteEnv",
	}

	assert.True(t, serviceEndPoint.IsEmpty())
}
