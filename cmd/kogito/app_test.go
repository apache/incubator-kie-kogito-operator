package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mitchellh/go-homedir"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAppCmd_WhenTheresNoConfigAndNoNamespace(t *testing.T) {
	home, _ := homedir.Dir()
	ns := uuid.New().String()
	path := filepath.Join(home, defaultConfigPath, defaultConfigFinalName)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		err := os.Remove(path)
		assert.NoError(t, err)
	} else {
		err := os.MkdirAll(filepath.Join(home, defaultConfigPath), os.ModePerm)
		assert.NoError(t, err)
	}

	// open
	file, err := os.Create(path)
	defer file.Close()
	assert.NoError(t, err)

	_, _, err = executeCli(strings.Join([]string{"app", ns}, " "))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), ns)
}

func TestAppCmd_WhenThereIsTheNamespace(t *testing.T) {
	ns := uuid.New().String()
	nsObj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	o, _, err := executeCli(strings.Join([]string{"app", ns}, " "), nsObj)
	assert.NoError(t, err)
	assert.NotEmpty(t, o)
	assert.Contains(t, o, ns)
}

func TestAppCmd_WhenWhatIsTheNamespace(t *testing.T) {
	ns := uuid.New().String()
	nsObj := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	// set the namespace
	executeCli(strings.Join([]string{"app", ns}, " "), nsObj)
	o, _, err := executeCli("app")
	assert.NoError(t, err)
	assert.NotEmpty(t, o)
	assert.Contains(t, o, ns)
}
