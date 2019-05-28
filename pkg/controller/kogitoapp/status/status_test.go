package status

import (
	"fmt"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetDeployed(t *testing.T) {
	now := metav1.Now()
	cr := &v1alpha1.KogitoApp{}

	assert.True(t, SetDeployed(cr))

	assert.NotEmpty(t, cr.Status.Conditions)
	assert.Equal(t, v1alpha1.DeployedConditionType, cr.Status.Conditions[0].Type)
	assert.Equal(t, corev1.ConditionTrue, cr.Status.Conditions[0].Status)
	assert.True(t, now.Before(&cr.Status.Conditions[0].LastTransitionTime))
}

func TestSetDeployedSkipUpdate(t *testing.T) {
	cr := &v1alpha1.KogitoApp{}
	SetDeployed(cr)

	assert.NotEmpty(t, cr.Status.Conditions)
	condition := cr.Status.Conditions[0]

	assert.False(t, SetDeployed(cr))
	assert.Equal(t, 1, len(cr.Status.Conditions))
	assert.Equal(t, condition, cr.Status.Conditions[0])
}

func TestSetProvisioning(t *testing.T) {
	now := metav1.Now()
	cr := &v1alpha1.KogitoApp{}
	assert.True(t, SetProvisioning(cr))

	assert.NotEmpty(t, cr.Status.Conditions)
	assert.Equal(t, v1alpha1.ProvisioningConditionType, cr.Status.Conditions[0].Type)
	assert.Equal(t, corev1.ConditionTrue, cr.Status.Conditions[0].Status)
	assert.True(t, now.Before(&cr.Status.Conditions[0].LastTransitionTime))
}

func TestSetProvisioningSkipUpdate(t *testing.T) {
	cr := &v1alpha1.KogitoApp{}
	assert.True(t, SetProvisioning(cr))

	assert.NotEmpty(t, cr.Status.Conditions)
	condition := cr.Status.Conditions[0]

	assert.False(t, SetProvisioning(cr))
	assert.Equal(t, 1, len(cr.Status.Conditions))
	assert.Equal(t, condition, cr.Status.Conditions[0])
}

func TestSetProvisioningAndThenDeployed(t *testing.T) {
	now := metav1.Now()
	cr := &v1alpha1.KogitoApp{}

	assert.True(t, SetProvisioning(cr))
	assert.True(t, SetDeployed(cr))

	assert.NotEmpty(t, cr.Status.Conditions)
	condition := cr.Status.Conditions[0]
	assert.Equal(t, 2, len(cr.Status.Conditions))
	assert.Equal(t, v1alpha1.ProvisioningConditionType, condition.Type)
	assert.Equal(t, corev1.ConditionTrue, condition.Status)
	assert.True(t, now.Before(&condition.LastTransitionTime))

	assert.Equal(t, v1alpha1.DeployedConditionType, cr.Status.Conditions[1].Type)
	assert.Equal(t, corev1.ConditionTrue, cr.Status.Conditions[1].Status)
	assert.True(t, condition.LastTransitionTime.Before(&cr.Status.Conditions[1].LastTransitionTime))
}

func TestBuffer(t *testing.T) {
	cr := &v1alpha1.KogitoApp{}
	for i := 0; i < maxBuffer+2; i++ {
		SetFailed(cr, v1alpha1.UnknownReason, fmt.Errorf("Error %d", i))
	}
	size := len(cr.Status.Conditions)
	assert.Equal(t, maxBuffer, size)
	assert.Equal(t, "Error 31", cr.Status.Conditions[size-1].Message)
}
