package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_WhenThereAreNoOperator(t *testing.T) {
	cli := fmt.Sprintf("deploy example-drools https://github.com/kiegroup/kogito-examples --context drools-quarkus-example --app kogito")
	lines, _, err := executeCli(cli, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kogito"}})
	assert.Error(t, err)
	assert.Contains(t, lines, "Couldn't find the Kogito Operator")
}
