package meta

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ResourceObject is any object in a Kubernetes or OpenShift cluster
type ResourceObject interface {
	metav1.Object
	runtime.Object
}
