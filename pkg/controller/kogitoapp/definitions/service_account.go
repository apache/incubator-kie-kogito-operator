package definitions

import (
	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ServiceAccountName default name for the SA responsible for running on Kogito App
	ServiceAccountName = "kogito-service"
)

// NewServiceAccount creates the ServiceAccount resource structure for the KogitoApp
func NewServiceAccount(kogitoApp *v1alpha1.KogitoApp) (serviceAccount corev1.ServiceAccount) {
	serviceAccount = corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceAccountName,
			Namespace: kogitoApp.Namespace,
		},
	}
	setGroupVersionKind(&serviceAccount.TypeMeta, ServiceAccountKind)
	addDefaultMeta(&serviceAccount.ObjectMeta, kogitoApp)
	return serviceAccount
}
