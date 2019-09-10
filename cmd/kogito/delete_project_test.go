package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_DeleteProjectCmd_WhenWeSuccessfullyDelete(t *testing.T) {
	cli := fmt.Sprintf("delete-project kogito")
	lines, _, err := executeCli(cli,
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kogito"}})
	assert.NoError(t, err)
	assert.Contains(t, lines, "Successfully deleted Kogito Project kogito")
}

func Test_DeleteProjectCmd_WhenProjectDoesNotExist(t *testing.T) {
	cli := fmt.Sprintf("delete-project kogito")
	lines, _, err := executeCli(cli)
	assert.Error(t, err)
	assert.Contains(t, lines, "Project kogito not found")
}
