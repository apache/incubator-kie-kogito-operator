package definitions

import (
	"fmt"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultRoleName = "view"
	defaultRoleType = "Role"
)

// NewRoleBinding creates the RoleBinding definition for the KogitoApp that will be bound to the Kogito ServiceAccount
func NewRoleBinding(kogitoApp *v1alpha1.KogitoApp, serviceAccount *corev1.ServiceAccount) (roleBinding rbacv1.RoleBinding) {
	roleBinding = rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: serviceAccount.Namespace,
			Name:      fmt.Sprintf("%s-%s", ServiceAccountName, defaultRoleName),
		},
		RoleRef: rbacv1.RoleRef{
			Kind: defaultRoleType,
			Name: defaultRoleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      KindServiceAccount.Name,
				Namespace: serviceAccount.Namespace,
				Name:      serviceAccount.Name,
			},
		},
	}
	SetGroupVersionKind(&roleBinding.TypeMeta, KindRoleBinding)
	addDefaultMeta(&roleBinding.ObjectMeta, kogitoApp)
	return roleBinding
}
