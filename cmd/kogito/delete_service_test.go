package main

import (
	"fmt"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_DeleteServiceCmd_WhenWeSuccessfullyDelete(t *testing.T) {
	cli := fmt.Sprintf("delete-service example-drools --project kogito")
	lines, _, err := executeCli(cli,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kogito"}},
		&v1alpha1.KogitoApp{ObjectMeta: metav1.ObjectMeta{Name: "example-drools", Namespace: "kogito"}})
	assert.NoError(t, err)
	assert.Contains(t, lines, "Successfully deleted Kogito Service example-drools")
}

func Test_DeleteServiceCmd_WhenServiceDoesNotExist(t *testing.T) {
	cli := fmt.Sprintf("delete-service example-drools --project kogito")
	lines, _, err := executeCli(cli,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kogito"}})
	assert.Error(t, err)
	assert.Contains(t, lines, "with the name 'example-drools' doesn't exist")
}
