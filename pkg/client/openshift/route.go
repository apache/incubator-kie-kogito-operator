package openshift

import (
	"time"

	routev1 "github.com/openshift/api/route/v1"

	"k8s.io/apimachinery/pkg/types"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
)

// RouteInterface exposes common operations for Route API
type RouteInterface interface {
	GetHostFromRoute(routeKey types.NamespacedName) (string, error)
}

func newRoute(c *client.Client) RouteInterface {
	client.MustEnsureClient(c)
	return &route{
		client: c,
	}
}

type route struct {
	client *client.Client
}

func (r *route) GetHostFromRoute(routeKey types.NamespacedName) (string, error) {
	route := &routev1.Route{}

	for i := 1; i < 60; i++ {
		time.Sleep(time.Duration(100) * time.Millisecond)
		if exists, err :=
			kubernetes.ResourceC(r.client).FetchWithKey(routeKey, route); exists {
			break
		} else if err != nil {
			log.Error("Error getting Route. ", err)
			return "", err
		}
	}

	return route.Spec.Host, nil
}
