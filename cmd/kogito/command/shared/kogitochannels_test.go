package shared

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_IsValidChannel(t *testing.T) {
	ch := "alpha"
	err := IsValidChannel(ch)
	assert.NoError(t, err)
}

func Test_IsValidChannelForInvalidInput(t *testing.T) {
	ch := "testChannel"
	err := IsValidChannel(ch)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Kogito channel type : testChannel")
}
