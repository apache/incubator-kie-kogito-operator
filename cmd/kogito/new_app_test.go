package main

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

func TestNewApp_WhenNewAppDoesNotExist(t *testing.T) {
	cli := fmt.Sprintf("new-app --name test")
	lines, _, err := executeCli(cli)
	assert.NoError(t, err)
	assert.Contains(t, lines, "created")
}

func TestNewApp_WhenNewAppExist(t *testing.T) {
	cli := fmt.Sprintf("new-app --name test")
	lines, _, err := executeCli(cli, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test"}})
	assert.NoError(t, err)
	assert.Contains(t, lines, "exists")
}

func TestNewApp_WhenTheresNoNamedFlag(t *testing.T) {
	cli := fmt.Sprintf("new-app test1")
	lines, _, err := executeCli(cli)
	assert.NoError(t, err)
	assert.Contains(t, lines, "created")
}
