package builder

import (
	"fmt"

	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultRoleName = "service-view"
	defaultRoleType = "Role"
)

// NewRole will create a namespaced custom role for the KogitoApp
func NewRole(kogitoApp *v1alpha1.KogitoApp) (role rbacv1.Role) {
	role = rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: kogitoApp.Namespace,
			Name:      defaultRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"list", "get", "watch"},
			},
		},
	}
	meta.SetGroupVersionKind(&role.TypeMeta, meta.KindRoleBinding)
	addDefaultMeta(&role.ObjectMeta, kogitoApp)
	return role
}

// NewRoleBinding creates the RoleBinding definition for the KogitoApp that will be bound to the Kogito ServiceAccount
func NewRoleBinding(kogitoApp *v1alpha1.KogitoApp, serviceAccount *corev1.ServiceAccount, role *rbacv1.Role) (roleBinding rbacv1.RoleBinding) {
	roleBinding = rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: kogitoApp.Namespace,
			Name:      fmt.Sprintf("%s-%s", ServiceAccountName, defaultRoleName),
		},
		RoleRef: rbacv1.RoleRef{
			Kind: defaultRoleType,
			Name: role.Name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      meta.KindServiceAccount.Name,
				Namespace: serviceAccount.Namespace,
				Name:      serviceAccount.Name,
			},
		},
	}
	meta.SetGroupVersionKind(&roleBinding.TypeMeta, meta.KindRoleBinding)
	addDefaultMeta(&roleBinding.ObjectMeta, kogitoApp)
	return roleBinding
}
